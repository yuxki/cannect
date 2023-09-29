package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/yuxki/cannect/pkg/asset"
	catalogapi "github.com/yuxki/cannect/pkg/catalog"
	orderapi "github.com/yuxki/cannect/pkg/order"
	uriapi "github.com/yuxki/cannect/pkg/uri"
)

type CatalogJSON struct {
	Alias    string `json:"alias"`
	URI      string `json:"uri"`
	Category string `json:"category"`
}

type OrderJSON struct {
	CatalogAliases []string `json:"aliases"`
	URI            string   `json:"uri"`
}

type CatalogsJSON struct {
	Catalogs []CatalogJSON `json:"catalogs"`
}

type OrdersJSON struct {
	Orders []OrderJSON `json:"orders"`
}

type CAnnectJSON struct {
	Catalogs []CatalogJSON `json:"catalogs"`
	Orders   []OrderJSON   `json:"orders"`
}

type runConfig struct {
	EnvOut   string
	ConLimit int
}

// Order is a struct that retrieves data from its own catalog and writes the
// contents at the destination specified by its own URI.
type Order interface {
	Order(context.Context) error
}

func newRunConfig(envOut string, conLimit int) runConfig {
	return runConfig{
		EnvOut:   envOut,
		ConLimit: conLimit,
	}
}

type catalogLogger struct {
	l *log.Logger
}

func (c *catalogLogger) Log(uriText string) {
	c.l.Printf("Fetching: %s", uriText)
}

func createCatalogSets(cntJSON CAnnectJSON, logger *log.Logger) ([][]orderapi.Catalog, error) {
	catalogSets := make([][]orderapi.Catalog, 0, len(cntJSON.Orders))

	srcSchemeReg := regexp.MustCompile("^(file|github)")
	cLogger := catalogLogger{l: logger}

	orderJSONs := cntJSON.Orders
	for idx := range orderJSONs {
		catalogSet := make([]orderapi.Catalog, 0, len(orderJSONs[idx].CatalogAliases))
		aliases := orderJSONs[idx].CatalogAliases
		for aliasIdx := range aliases {
			var cJSON CatalogJSON
			var ok bool

			for _, jsn := range cntJSON.Catalogs {
				if jsn.Alias == aliases[aliasIdx] {
					cJSON = jsn
					ok = true
					break
				}
			}

			if !ok {
				panic(fmt.Sprintf("alias in destination not found in sources: %s", aliases[aliasIdx]))
			}

			var checker catalogapi.AssetChecker

			switch cJSON.Category {
			case asset.CertCategory:
				checker = asset.NewCertiricate()
			case asset.PrivKeyCategory:
				checker = asset.NewPrivateKey()
			case asset.EncPrivKeyCategory:
				checker = asset.NewEncryptedPrivateKey()
			case asset.CRLCategory:
				checker = asset.NewCRL()
			default:
				panic(fmt.Sprintf("Undefined category: %s", cJSON.Category))
			}

			var catalog orderapi.Catalog
			scheme := srcSchemeReg.FindString(cJSON.URI)

			switch scheme {
			case "file":
				uri, err := uriapi.NewFSURI(cJSON.URI)
				if err != nil {
					return nil, err
				}
				catalog = catalogapi.NewFSCatalog(uri, cJSON.Alias, checker).WithLogger(&cLogger)
			case "github":
				uri, err := uriapi.NewGitHubURI(cJSON.URI)
				if err != nil {
					return nil, err
				}
				catalog = catalogapi.NewGitHubCatalog(uri, cJSON.Alias, checker).WithLogger(&cLogger)
			default:
				panic(fmt.Sprintf("Undefined source storage: %s", scheme))
			}

			catalogSet = append(catalogSet, catalog)
		}
		catalogSets = append(catalogSets, catalogSet)
	}

	return catalogSets, nil
}

type orderLogger struct {
	l *log.Logger
}

func (o *orderLogger) Log(uriText string) {
	o.l.Printf("Ordering: %s", uriText)
}

func run(ctx context.Context, cntJSON CAnnectJSON, cfg runConfig, logger *log.Logger) error {
	catalogSets, err := createCatalogSets(cntJSON, logger)
	if err != nil {
		return err
	}

	// Order to destinations
	var envFile *os.File
	var wg sync.WaitGroup
	limit := make(chan struct{}, cfg.ConLimit)

	dstSchemeReg := regexp.MustCompile("^(file|env)")

	oLog := orderLogger{l: logger}

	for idx, oJSON := range cntJSON.Orders {
		idx := idx
		wg.Add(1)

		var order Order
		scheme := dstSchemeReg.FindString(oJSON.URI)

		switch scheme {
		case "file":
			uri, err := uriapi.NewFSURI(oJSON.URI)
			if err != nil {
				return err
			}

			order = orderapi.NewFSOrder(uri, catalogSets[idx]).WithLogger(&oLog)
		case "env":
			uri, err := uriapi.NewEnvURI(oJSON.URI)
			if err != nil {
				return err
			}

			if envFile == nil {
				envFile, err = os.Create(cfg.EnvOut)
				if err != nil {
					return err
				}
				defer envFile.Close()
			}

			order = orderapi.NewEnvOrder(uri, catalogSets[idx], envFile).WithLogger(&oLog)
		default:
			panic(fmt.Sprintf("Undefined destination scheme: %s", scheme))
		}

		go func() {
			limit <- struct{}{}
			err := order.Order(ctx)
			if err != nil {
				panic(err)
			}

			wg.Done()
			<-limit
		}()
	}

	wg.Wait()

	return nil
}

func unmarshal(file *os.File) CAnnectJSON {
	var jsn CAnnectJSON
	err := json.NewDecoder(file).Decode(&jsn)
	if err != nil {
		panic(err)
	}

	return jsn
}

func unmarshalBoth(cFile, oFile *os.File) CAnnectJSON {
	var cJSON CatalogsJSON
	err := json.NewDecoder(cFile).Decode(&cJSON)
	if err != nil {
		panic(err)
	}

	var oJSON OrdersJSON
	err = json.NewDecoder(oFile).Decode(&oJSON)
	if err != nil {
		panic(err)
	}

	return CAnnectJSON{
		Catalogs: cJSON.Catalogs,
		Orders:   oJSON.Orders,
	}
}

type InvalidOrderFileError struct {
	reason string
}

func (e InvalidOrderFileError) Error() string {
	return fmt.Sprintf("invalid order file: %s", e.reason)
}

func validate(jsn CAnnectJSON) error {
	alsMap := make(map[string]interface{})
	for i := range jsn.Catalogs {
		alsMap[jsn.Catalogs[i].Alias] = nil
	}

	dupMap := make(map[string]interface{})
	oJSONs := jsn.Orders
	for idx := range oJSONs {
		aliases := oJSONs[idx].CatalogAliases
		for _, als := range aliases {
			_, ok := alsMap[als]
			if !ok {
				// Check no undefined alias
				return InvalidOrderFileError{
					reason: fmt.Sprintf("undefined aliase: %s", als),
				}
			}
		}

		if _, ok := dupMap[oJSONs[idx].URI]; !ok {
			dupMap[oJSONs[idx].URI] = nil
			continue
		}
		// Check No Duplicated destination
		return InvalidOrderFileError{
			reason: fmt.Sprintf("Order URI must not be duplicated: %s", oJSONs[idx].URI),
		}
	}

	return nil
}

const (
	defaultTimeout  = 30
	defaultEnvOut   = "./cannect.env"
	defaultConLimit = 5
)

const (
	catalogFlg      = 0x01
	orderFlg        = 0x02
	catalogOrderFlg = 0x04
)

func checkExclusive(catalog, order, catalogOrder string) (int, bool) {
	flgs := 0x00

	if len(catalog) > 0 {
		flgs |= catalogFlg
	}

	if len(order) > 0 {
		flgs |= orderFlg
	}

	if len(catalogOrder) > 0 {
		flgs |= catalogOrderFlg
	}

	switch flgs {
	case catalogFlg | orderFlg:
		return flgs, true
	case catalogOrderFlg:
		return flgs, true
	}

	return flgs, false
}

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags)

	catalog := flag.String("catalog", "", "The path of JSON format file contains catalogs.")
	order := flag.String("order", "", "The path of JSON format file contains orders.")
	catalogOrder := flag.String("catalog-order", "", "The path of JSON format file contains catalogs and orders.")
	envOut := flag.String("env-out", defaultEnvOut, "'env' scheme output file.")
	conLimit := flag.Int("con-limit", defaultConLimit, "The limit of concurrency..")
	timeout := flag.Int64("timeout", defaultTimeout, "Timeout (seconds).")
	flag.Parse()

	flgs, ok := checkExclusive(*catalog, *order, *catalogOrder)
	if !ok {
		// nolint lll
		logger.Fatalln(
			`
Usage: cannect <OPTIONS>
  OPTIONS
    -catalog <file-path> The path of catalog file. (required: Exclusive to -catalog-order)
    -order <file-path> The path of order file. (required: Exclusive to -catalog-order)
    -catalog-order <file-path> The path of file contains both orders and catalogs. (required: Exclusive to -catalog and -order)
    -env-out <file-path> The path of env scheme output. (default: ./cannect.env)
    -con-limit <number> The limit of concurrency. (default: 5)
    -timeout <number> The number of seconds for timeout. (default: 30)`,
		)
	}

	var cntJSON CAnnectJSON
	switch flgs {
	case catalogFlg | orderFlg:
		cFile, err := os.Open(*catalog)
		if err != nil {
			log.Fatalln(err)
		}

		oFile, err := os.Open(*order)
		if err != nil {
			log.Fatalln(err)
		}

		cntJSON = unmarshalBoth(cFile, oFile)
	case catalogOrderFlg:
		file, err := os.Open(*catalogOrder)
		if err != nil {
			log.Fatalln(err)
		}

		cntJSON = unmarshal(file)
	}

	err := validate(cntJSON)
	if err != nil {
		log.Fatalln(err)
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(*timeout))
	defer cancel()

	cfg := newRunConfig(*envOut, *conLimit)
	err = run(ctx, cntJSON, cfg, logger)
	if err != nil {
		log.Println(err)
	}
}
