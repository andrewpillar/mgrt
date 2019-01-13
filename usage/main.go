package usage

var (
	Commands = map[string]string{
		"add":   Add,
		"init":  Init,
		"log":   Log,
		"ls":    Ls,
		"run":   Run,
		"reset": Reset,
	}

	Main = `mgrt - Simple SQL migrations

Usage:

  mgrt [command] [options...]

Commands:

  add    Add a new revision
  init   Initialize a new mgrt instance
  log    Display performed revisions
  ls     List available revisions
  run    Run a revision
  reset  Reset an already run revision

Options:

  --help  Display this usage message

For more information on a command run 'mgrt [command] --help'`
)
