package managerservice

import (
	"neko-manager/pkg/instancerepo"
	"neko-manager/pkg/nekosupplier"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/teadove/teasutils/utils/test_utils"
)

func TestRequireRegularReport(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		Name      string
		CreatedAt time.Time
		UpdatedAt time.Time
		Exp       bool
	}{
		{
			Name:      "hour before",
			CreatedAt: now.Add(-time.Hour),
			UpdatedAt: now.Add(-time.Hour),
			Exp:       true,
		},
		{
			Name:      "just updated",
			CreatedAt: now.Add(-time.Hour),
			UpdatedAt: now,
			Exp:       false,
		},
		{
			Name:      "just created",
			CreatedAt: now,
			UpdatedAt: now.Add(-time.Hour),
			Exp:       false,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(tt *testing.T) {
			tt.Parallel()

			act := requireRegularReport(&instancerepo.Instance{
				CreatedAt: test.CreatedAt,
				UpdatedAt: test.UpdatedAt,
			})

			assert.Equal(tt, test.Exp, act)
		})
	}
}

func TestRequireDeletion(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		Name  string
		Stats nekosupplier.Stats
		Exp   bool
	}{
		{
			Name:  "with users",
			Stats: nekosupplier.Stats{TotalUsers: 1},
			Exp:   false,
		},
		{
			Name:  "with admins",
			Stats: nekosupplier.Stats{TotalAdmins: 1},
			Exp:   false,
		},
		{
			Name:  "user just left",
			Stats: nekosupplier.Stats{LastUserLeftAt: now},
			Exp:   false,
		},
		{
			Name:  "admin just left",
			Stats: nekosupplier.Stats{LastAdminLeftAt: now},
			Exp:   false,
		},
		{
			Name:  "server just created",
			Stats: nekosupplier.Stats{ServerStartedAt: now},
			Exp:   false,
		},
		{
			Name: "not used for long",
			Stats: nekosupplier.Stats{
				ServerStartedAt: now.Add(-time.Hour),
				LastAdminLeftAt: now.Add(-time.Hour),
				LastUserLeftAt:  now.Add(-time.Hour),
			},
			Exp: true,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(tt *testing.T) {
			tt.Parallel()

			act := requireDeletion(test_utils.GetLoggedContext(), &test.Stats)
			assert.Equal(tt, test.Exp, act)
		})
	}
}
