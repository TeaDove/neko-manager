package instancerepo

import (
	"crypto/rand"
	"neko-manager/pkg/nekosupplier"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepr(t *testing.T) {
	t.Parallel()

	instance := Instance{
		ID:              "gyqlvj",
		Status:          InstanceStatusRunning,
		CreatedAt:       time.Now().Add(-15 * time.Minute),
		UpdatedAt:       time.Now(),
		CreatedBy:       "TeaDove",
		TGChatID:        418878871,
		SessionAPIToken: rand.Text(),
		CloudFolderID:   "b1gt2lbgae1f073bjo0u",
		CloudInstanceID: "epdec5ei91e5aeg732ok",
		IP:              "158.160.84.42",
		ProxyURL:        "https://kodiki-hack.ru:8080",
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

	//nolint: lll // Is string
	assert.Equal(t, `🐈‍⬛ Neko instance &lt;<code>gyqlvj</code>&gt; (@TeaDove)
<b>✅ Running ✅</b>
Alive for 15m

User login: https://kodiki-hack.ru:8080?pwd=neko
Admin login: https://kodiki-hack.ru:8080?pwd=admin
<span class="tg-spoiler">User login unsecure: http://158.160.84.42?pwd=neko
Admin login unsecure: http://158.160.84.42?pwd=admin</span>
IP: 158.160.84.42
SSH: <code>ssh -oStrictHostKeyChecking=no -i ~/.ssh/id_rsa_yc -v neko@158.160.84.42</code>
Cloud: <a href="https://console.yandex.cloud/folders/b1gt2lbgae1f073bjo0u/compute/instance/epdec5ei91e5aeg732ok/overview">yc</a>

Load: not used for = 3m; host = <code>SOME-HOST</code>; total users = 3; last admin left = 3m ago`, repr)
}
