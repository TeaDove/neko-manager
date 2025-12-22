package instancerepo

import (
	"crypto/rand"
	"neko-manager/pkg/nekosupplier"
	"neko-manager/pkg/randutils"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepr(t *testing.T) {
	t.Parallel()

	instance := Instance{
		ID:              randutils.RandomString(6),
		Status:          InstanceStatusRunning,
		CreatedAt:       time.Now().Add(-15 * time.Minute),
		UpdatedAt:       time.Now(),
		CreatedBy:       "TeaDove",
		TGChatID:        418878871,
		SessionAPIToken: rand.Text(),
		CloudFolderID:   "b1gt2lbgae1f073bjo0u",
		CloudInstanceID: "epdec5ei91e5aeg732ok",
		IP:              "158.160.84.42",
	}

	repr, err := instance.Repr(&nekosupplier.Stats{
		HasHost:         true,
		HostId:          "SOME-HOST",
		ServerStartedAt: time.Now().Add(-10 * time.Minute),
		TotalUsers:      3,
		LastAdminLeftAt: time.Now().Add(-3 * time.Minute),
	})
	require.NoError(t, err)
	assert.NotEmpty(t, repr)

	println(repr)
}
