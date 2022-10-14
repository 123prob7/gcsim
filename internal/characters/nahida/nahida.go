package nahida

import (
	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/core/player/character/profile"
)

func init() {
	core.RegisterCharFunc(keys.Nahida, NewChar)
}

type elementsCounter struct {
	pyro    int
	electro int
	hydro   int
}

type char struct {
	*tmpl.Character
	triIcd            int
	triInterval       int
	markedTargetCount int
	qCounter          elementsCounter
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
	c.qCounter = elementsCounter{pyro: 0, electro: 0, hydro: 0}
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

	for _, this := range c.Core.Player.Chars() {
		switch this.Base.Element {
		case attributes.Pyro:
			c.qCounter.pyro++
		case attributes.Electro:
			c.qCounter.electro++
		case attributes.Hydro:
			c.qCounter.hydro++
		default:
		}
	}

	if c.Base.Cons >= 1 {
		c.qCounter.pyro++
		c.qCounter.electro++
		c.qCounter.hydro++
	}

	if c.qCounter.pyro > 2 {
		c.qCounter.pyro = 2
	}
	if c.qCounter.electro > 2 {
		c.qCounter.electro = 2
	}
	if c.qCounter.hydro > 2 {
		c.qCounter.hydro = 2
	}

	return nil
}
