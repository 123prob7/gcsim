package nahida

import (
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const a1EMLimit = 250

// TODO: recheck how a1 works (how it interacts with EM buffers). Assuming it just checks characters' stats onset
func (c *char) a1() {
	var emShare float64 = 0
	for _, this := range c.Core.Player.Chars() {
		if this.Stat(attributes.EM) > emShare {
			emShare = this.Stat(attributes.EM)
		}
	}
	emShare *= .25
	if emShare > a1EMLimit {
		emShare = a1EMLimit
	}

	for _, this := range c.Core.Player.Chars() {
		ind := this.Index
		this.AddStatMod(character.StatMod{
			Base:         modifier.NewBase("nahida-a1", -1),
			AffectedStat: attributes.EM,
			Amount: func() ([]float64, bool) {
				val := make([]float64, attributes.EndStatType)
				if c.Core.Status.Duration(burstKey) == 0 {
					return val, false
				}
				if ind != c.Core.Player.ActiveChar().Index {
					return val, false
				}

				val[attributes.EM] = emShare
				return val, true
			},
		})
	}
}

func (c *char) a4() {
	val := make([]float64, attributes.EndStatType)
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("nahida-a4", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			if atk.Info.AttackTag != combat.AttackTagElementalArt {
				return val, false
			}
			if atk.Info.Abil != "Tri-Karma Purification" && atk.Info.Abil != "Tri-Karma Purification (C6)" {
				return val, false
			}
			excess := int(c.Stat(attributes.EM) - 200)
			if excess < 0 {
				excess = 0
			}

			val[attributes.CR] = .03 / 100 * float64(excess)
			val[attributes.DmgP] = .1 / 100 * float64(excess)
			if val[attributes.CR] > .24 {
				val[attributes.CR] = .24
			}
			if val[attributes.DmgP] > .8 {
				val[attributes.DmgP] = .8
			}

			return val, true
		},
	})
}
