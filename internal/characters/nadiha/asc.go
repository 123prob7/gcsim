package nadiha

import (
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const a1EMLimit = 250

// TODO: recheck this a1 if it snapshots or not?
// It shouldnt snapshot. Bc % em buffs dont benefit from other % em buff sources.
func (c *char) a1() {
	c.a1UpdateEM()

	for _, this := range c.Core.Player.Chars() {
		ind := this.Index
		this.AddStatMod(character.StatMod{
			Base:         modifier.NewBase("nadiha-a1", -1),
			AffectedStat: attributes.EM,
			Amount: func() ([]float64, bool) {
				val := make([]float64, attributes.EndStatType)
				if c.Core.Status.Duration(burstKey) == 0 {
					return val, false
				}
				if ind != c.Core.Player.ActiveChar().Index {
					return val, false
				}

				val[attributes.EM] = c.highestEMShare
				return val, true
			},
		})
	}
}

func (c *char) a1UpdateEM() {
	for _, this := range c.Core.Player.Chars() {
		if this.Stat(attributes.EM) > c.highestEMShare {
			c.highestEMShare = this.Stat(attributes.EM)
		}
	}
	c.highestEMShare *= .25
	if c.highestEMShare > a1EMLimit {
		c.highestEMShare = a1EMLimit
	}
}

// // TODO: recheck this. Should it also work with em buffer??
// func (c *char) a1QueueUpdateEMTask(src int) func() {
// 	return func() {
// 		if src != c.a1Src {
// 			return
// 		}

// 		prev := c.highestEMShare
// 		c.a1UpdateEM()
// 		if prev != c.highestEMShare {
// 			c.Core.Log.NewEvent("nadiha a1 updating em", glog.LogCharacterEvent, c.Index).
// 				Write("prev", prev).
// 				Write("new", c.highestEMShare).
// 				Write("src", src)
// 		}

// 		// c.Core.Tasks.Add(c.a1QueueUpdateEMTask(src), 60)
// 	}
// }

func (c *char) a4() {
	val := make([]float64, attributes.EndStatType)
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("nadiha-a4", -1),
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
