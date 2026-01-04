package nekoproxy

import (
	"fmt"
	"neko-manager/pkg/randutils"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teadove/teasutils/utils/must_utils"
)

func TestGetTarget(t *testing.T) {
	t.Parallel()
	t.Skip("Потому что откатили фичу")

	idLen := 6
	r := New(idLen, "http://localhost:8080")

	id1 := randutils.RandomString(idLen)
	id2 := randutils.RandomString(idLen)

	r.AddTarget(id1, must_utils.Must(url.Parse("http://1.1.1.1")))
	r.AddTarget(id2, must_utils.Must(url.Parse("http://2.2.2.2")))

	assert.Equal(t,
		"https://google.com",
		r.getTarget(must_utils.Must(url.Parse("http://localhost:8080")).Path).String(),
	)
	assert.Equal(t,
		"https://google.com",
		r.getTarget(must_utils.Must(url.Parse("http://localhost:8080/")).Path).String(),
	)
	assert.Equal(t,
		"http://1.1.1.1",
		r.getTarget(must_utils.Must(url.Parse(fmt.Sprintf("http://localhost:8080/%s", id1))).Path).String(),
	)
	assert.Equal(t,
		"http://1.1.1.1",
		r.getTarget(must_utils.Must(url.Parse(fmt.Sprintf("http://localhost:8080/%s/", id1))).Path).String(),
	)
	assert.Equal(t,
		"http://1.1.1.1",
		r.getTarget(must_utils.Must(url.Parse(fmt.Sprintf("http://localhost:8080/%s/somePath", id1))).Path).String(),
	)
	assert.Equal(t,
		"http://2.2.2.2",
		r.getTarget(must_utils.Must(url.Parse(fmt.Sprintf("http://localhost:8080/%s/somePath", id2))).Path).String(),
	)

	r.DeleteTarget(id1)
	r.DeleteTarget(id2)
	assert.Equal(t,
		"https://google.com",
		r.getTarget(must_utils.Must(url.Parse(fmt.Sprintf("http://localhost:8080/%s/somePath", id1))).Path).String(),
	)
	assert.Equal(t,
		"https://google.com",
		r.getTarget(must_utils.Must(url.Parse(fmt.Sprintf("http://localhost:8080/%s/somePath", id2))).Path).String(),
	)
}
