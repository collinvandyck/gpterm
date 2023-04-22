package gptea

type SetCredentialReq struct {
	Prompt string
	Key    string
}

type SetCredentialRes struct {
	Err error
}
