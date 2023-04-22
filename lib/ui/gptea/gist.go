package gptea

type GistResultMsg struct {
	URL           string
	Opened        bool
	Err           error
	NoCredentials bool
}

type GistOpenReq struct {
	URL string
}

type GistOpenRes struct {
	URL string
	Err error
}
