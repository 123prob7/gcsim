package nadiha

import (
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/glog"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const nadihaC6Key = "nadihac6"

// Burning, Bloom, Hyperbloom, and Burgeon Reaction DMG can score CRIT Hits. CRIT Rate and CRIT DMG are fixed at 20% and 100% respectively
// Within 8s of being affected by Quicken, Aggravate, Spread, DEF is decreased by 30%
func (c *char) c2() {
	if c.Base.Cons < 2 {
		return
	}

	c.Core.Events.Subscribe(event.OnAttackWillLand, func(args ...interface{}) bool {
		ae := args[1].(*combat.AttackEvent)

		switch ae.Info.Abil {
		case string(combat.Burning), string(combat.Bloom), string(combat.Hyperbloom), string(combat.Burgeon):
			ae.Snapshot.Stats[attributes.CR] = .2
			ae.Snapshot.Stats[attributes.CD] = 1
		default:
		}

		return false
	}, "nadiha-c2")

	cb := func(args ...interface{}) bool {
		t := args[0].(combat.Target)
		e, ok := t.(*enemy.Enemy)
		if !ok {
			return false
		}

		e.AddDefMod(enemy.DefMod{
			Base:  modifier.NewBaseWithHitlag("nadiha-c2", 8*60),
			Value: -0.3,
		})
		return false
	}

	c.Core.Events.Subscribe(event.OnQuicken, cb, "nadiha-c2")
	c.Core.Events.Subscribe(event.OnAggravate, cb, "nadiha-c2")
	c.Core.Events.Subscribe(event.OnSpread, cb, "nadiha-c2")
}

// C4: The Stem of Manifest Inference
// When 1/2/3/(4 or more) nearby opponents are affected by All Schemes to Know's Seeds of Skandha,
// Nahida's Elemental Mastery will be increased by 100/120/140/160.
func (c *char) c4() {
	if c.Base.Cons < 4 {
		return
	}

	c.AddStatMod(character.StatMod{
		Base: modifier.NewBase("nadiha-c4", -1),
		Amount: func() ([]float64, bool) {
			val := make([]float64, attributes.EndStatType)

			if c.markedTargetCount > 0 {
				lim := c.markedTargetCount
				if lim > 4 {
					lim = 4
				}
				val[attributes.EM] = float64(80 + 20*lim)
			}
			return val, true
		},
	})
}

// C6: The Fruit of Reason's Culmination
func (c *char) c6Cb(a combat.AttackCB) {
	if c.Base.Cons < 6 {
		return
	}
	switch a.AttackEvent.Info.AttackTag {
	case combat.AttackTagNormal, combat.AttackTagExtra:
	default:
		return
	}
	// in icd
	if c.c6Icd > c.Core.F {
		return
	}
	// ran out of stack
	if c.c6Stacks == 0 {
		c.DeleteStatus(nadihaC6Key)
		return
	}
	// status expired, no need to check nadihaburst field
	if !c.StatusIsActive(nadihaC6Key) {
		return
	}

	c.c6Icd = c.Core.F + .2*60
	c.c6Stacks--

	pyroBonus := burstPyroBonus[c.pyroCount][c.TalentLvlBurst()]
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Tri-Karma Purification (C6)",
		AttackTag:  combat.AttackTagElementalArt,
		ICDTag:     combat.ICDTagElementalArt,
		ICDGroup:   combat.ICDGroupDefault,
		Element:    attributes.Dendro,
		Durability: 37.5,
		Mult:       2 + pyroBonus,
	}
	ai.FlatDmg = 4 * c.Stat(attributes.EM)

	// queue dmg to all marked enemies
	for _, e := range c.Core.Combat.Enemies() {
		if !targetHasSkandhaKey(e) {
			continue
		}
		c.Core.QueueAttack(ai, combat.NewDefSingleTarget(e.Key(), combat.TargettableEnemy), 5, 5)
	}

	c.Core.Log.NewEvent("Tri-Karma C6 ticking", glog.LogCharacterEvent, c.Index).
		Write("next", c.c6Icd).
		Write("stacksLeft", c.c6Stacks).
		Write("pyroBonus", pyroBonus).
		Write("src", a.AttackEvent.Info.Abil)

	if c.triParticleIcd < c.Core.F {
		c.triParticleIcd = c.Core.F + 7.5*60
		c.Core.QueueParticle("nadiha", 3, attributes.Dendro, 80)
	}
}
