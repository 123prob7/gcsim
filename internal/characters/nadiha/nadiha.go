package nadiha

import (
	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/core/player/character/profile"
)

func init() {
	core.RegisterCharFunc(keys.Nadiha, NewChar)
}

type char struct {
	*tmpl.Character
	triIcd            int
	triInterval       int
	markedTargetCount int
	pyroCount         int
	electroCount      int
	hydroCount        int
	triParticleIcd    int
	c6Stacks          int
	c6Icd             int
}

func NewChar(s *core.Core, w *character.CharWrapper, _ profile.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.EnergyMax = 50
	c.NormalHitNum = normalHitNum
	c.SkillCon = 3
	c.BurstCon = 5

	c.triIcd = 0
	c.triInterval = 2.5 * 60
	c.markedTargetCount = 0
	c.pyroCount = 0
	c.electroCount = 0
	c.hydroCount = 0
	c.triParticleIcd = 0

	if c.Base.Cons >= 6 {
		c.c6Stacks = 6
		c.c6Icd = 0
	}

	w.Character = &c

	return nil
}

func (c *char) Init() error {
	c.triKarmaTickOnReaction()
	c.a1()
	c.a4()
	c.c2()

	return nil
}
