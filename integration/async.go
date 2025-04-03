package integration

type Async interface {
	Execute(request)
}

type asyncClient struct {
	tries int
	db    string
}

func (a *asyncClient) Execute(r request) {
	go func() {
		client := NewHttpClient[RawData]()
		for range a.tries {
			resp, err := client.Send(r)
			_ = resp
			_ = err
		}
	}()
}
