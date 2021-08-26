package iris_recorder_bug

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"github.com/kataras/iris/v12/core/host"
	"gotest.tools/v3/assert"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"testing"
)

func readBody(res *http.Response) ([]byte, error) {
	var reader io.Reader = res.Body
	if res.ContentLength <= 0 {
		reader = httputil.NewChunkedReader(reader)
	}
	return ioutil.ReadAll(reader)
}

func recorder(ctx context.Context) {
	if _, ok := ctx.IsRecording(); !ok {
		ctx.Record()
	}
}

func TestRecorderBug(t *testing.T) {
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		chunkedWriter := httputil.NewChunkedWriter(writer)
		writer.Header().Set("transfer-Encoding", "chunked")
		chunkedWriter.Write([]byte("hello"))
		chunkedWriter.Close()
	})

	go http.ListenAndServe(":9394", nil)

	res, err := http.Get("http://localhost:9394")
	assert.NilError(t, err)
	body, err := readBody(res)
	assert.NilError(t, err)
	assert.Equal(t, string(body), "hello")

	app := iris.New()
	//app.Use(recorder)
	u, err := url.Parse("http://localhost:9394")
	app.Get("/", iris.FromStd(host.ProxyHandler(u)))

	go app.Run(iris.Addr(":9395"))

	res, err = http.Get("http://localhost:9395")
	assert.NilError(t, err)
	body, err = readBody(res)
	assert.NilError(t, err, string(body))
	assert.Equal(t, string(body), "hello")
}
