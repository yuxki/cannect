package cannect

type Logger interface {
	// Log about provided uri.
	Log(uri URI)
}
