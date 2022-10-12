package nadiha

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/glog"
	"github.com/genshinsim/gcsim/pkg/enemy"
)

var (
	skillFrames     []int
	skillHoldFrames []int
)

// TODO: check frame
func init() {
	skillFrames = frames.InitAbilSlice(24)
	skillFrames[action.ActionDash] = 24
	skillFrames[action.ActionJump] = 24

	skillHoldFrames = frames.InitAbilSlice(62)
	skillHoldFrames[action.ActionDash] = 62
	skillHoldFrames[action.ActionJump] = 62
}

const (
	skillHitmark     = 8 // TODO: check frame
	seedOfSkandhaKey = "skandha"
	triKarmaDuration = 25 * 60
	maxMarkedTargets = 8
)

// Skill handling
func (c *char) Skill(p map[string]int) action.ActionInfo {
	hold, ok := p["hold"]
	if ok && (hold < 0 || hold > 1) {
		hold = 1
	}

	t, ok := p["targets"]
	if !ok || t > maxMarkedTargets {
		t = maxMarkedTargets
	}

	// recount marked targets
	c.markedTargetCount = 0
	for _, t := range c.Core.Combat.Enemies() {
		if targetHasSkandhaKey(t) {
			c.markedTargetCount++
		}
	}

	c.c4()

	if hold == 1 {
		return c.skillHold()
	}
	return c.skillPress()
}

// Skill press handling
func (c *char) skillPress() action.ActionInfo {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "All Schemes to Know (Press)",
		AttackTag:  combat.AttackTagElementalArt,
		ICDTag:     combat.ICDTagNone,
		ICDGroup:   combat.ICDGroupDefault,
		Element:    attributes.Dendro,
		Durability: 25,
		Mult:       skillDmg[c.TalentLvlSkill()],
	}
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHit(c.Core.Combat.Player(), 4, false, combat.TargettableEnemy),
		skillHitmark,
		skillHitmark,
		c.applyMarkCB,
	)

	// TODO: recheck frame
	c.SetCDWithDelay(action.ActionSkill, 5*60, 8)

	return action.ActionInfo{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionJump],
		State:           action.SkillState,
	}
}

// Skill hold handling
func (c *char) skillHold() action.ActionInfo {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "All Schemes to Know (Hold)",
		AttackTag:  combat.AttackTagElementalArtHold,
		ICDTag:     combat.ICDTagNone,
		ICDGroup:   combat.ICDGroupDefault,
		Element:    attributes.Dendro,
		Durability: 25,
		Mult:       skillHoldDmg[c.TalentLvlSkill()],
	}
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHit(c.Core.Combat.Player(), 8, false, combat.TargettableEnemy),
		skillHitmark,
		skillHitmark,
		c.applyMarkCB,
	)

	// TODO: recheck frame
	c.SetCDWithDelay(action.ActionSkill, 6*60, 8)

	return action.ActionInfo{
		Frames:          frames.NewAbilFunc(skillHoldFrames),
		AnimationLength: skillHoldFrames[action.InvalidAction],
		CanQueueAfter:   skillHoldFrames[action.ActionJump],
		State:           action.SkillState,
	}
}

// apply mark
func (c *char) applyMarkCB(a combat.AttackCB) {
	t, ok := a.Target.(*enemy.Enemy)
	if !ok {
		return
	}
	if c.markedTargetCount >= maxMarkedTargets {
		return
	}

	t.AddStatus(seedOfSkandhaKey, triKarmaDuration, true)
	c.markedTargetCount++
	c.Core.Log.NewEvent(
		"seed of skandha applied",
		glog.LogCharacterEvent,
		c.Index,
	).
		Write("target", t.Index()).
		Write("expiry", t.StatusExpiry(seedOfSkandhaKey))
}

// queue skill tick dmg to all enemies marked by nadiha skill
func (c *char) triKarmaTickOnReaction() {
	cb := func(args ...interface{}) bool {
		t := args[0].(combat.Target)
		ae := args[1].(*combat.AttackEvent)
		//ignore if targets not marked
		if !targetHasSkandhaKey(t) {
			return false
		}
		//ignore if skill tick on icd
		if c.triIcd > c.Core.F {
			return false
		}

		// in-burst bonus buffs
		var (
			pyroBonus    float64
			electroBonus int
		)
		if c.Core.Status.Duration(burstKey) > 0 {
			pyroBonus = burstPyroBonus[c.pyroCount][c.TalentLvlBurst()]
			electroBonus = int(burstElectroBonus[c.electroCount][c.TalentLvlBurst()] * 60)
		}

		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Tri-Karma Purification",
			AttackTag:  combat.AttackTagElementalArt,
			ICDTag:     combat.ICDTagNone,
			ICDGroup:   combat.ICDGroupDefault,
			Element:    attributes.Dendro,
			Durability: 37.5,
			Mult:       skillTickDmg[c.TalentLvlSkill()] + pyroBonus,
		}
		ai.FlatDmg = skillTickFlat[c.TalentLvlSkill()] * c.Stat(attributes.EM)

		c.triIcd = c.Core.F + c.triInterval - electroBonus

		// queue dmg to all marked enemies
		for _, e := range c.Core.Combat.Enemies() {
			if !targetHasSkandhaKey(e) {
				continue
			}
			c.Core.QueueAttack(ai, combat.NewDefSingleTarget(e.Key(), combat.TargettableEnemy), 5, 5)
		}

		// TODO: check frame, particle icd
		if c.triParticleIcd < c.Core.F {
			c.triParticleIcd = c.Core.F + 7.5*60
			c.Core.QueueParticle("nadiha", 3, attributes.Dendro, 80)
		}

		c.Core.Log.NewEvent("Tri-Karma ticking", glog.LogCharacterEvent, c.Index).
			Write("next", c.triIcd).
			Write("pyroBonus", pyroBonus).
			Write("electroBonus", electroBonus).
			Write("src", ae.Info.Abil)

		return false
	}

	for i := event.ReactionEventStartDelim + 1; i < event.ReactionEventEndDelim; i++ {
		c.Core.Events.Subscribe(i, cb, "nadiha-skill")
	}
}

// helper to check skandha key on target
func targetHasSkandhaKey(t combat.Target) bool {
	e, ok := t.(*enemy.Enemy)
	if !ok {
		return false
	}
	return e.StatusIsActive(seedOfSkandhaKey)
}
