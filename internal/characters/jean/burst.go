package jean

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/avatar"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/glog"
	"github.com/genshinsim/gcsim/pkg/core/player"
)

var burstFrames []int

const burstStart = 40

func init() {
	burstFrames = frames.InitAbilSlice(90) // Q -> D/J
	burstFrames[action.ActionAttack] = 88  // Q -> N1
	burstFrames[action.ActionSkill] = 89   // Q -> E
	burstFrames[action.ActionSwap] = 88    // Q -> Swap
}

func (c *char) Burst(p map[string]int) action.ActionInfo {
	// p is the number of times enemy enters or exits the field
	enter := p["enter"]
	if enter < 1 {
		enter = 1
	}
	delay, ok := p["enter_delay"]
	if !ok {
		delay = 600 / enter
	}

	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Dandelion Breeze",
		AttackTag:  combat.AttackTagElementalBurst,
		ICDTag:     combat.ICDTagNone,
		ICDGroup:   combat.ICDGroupDefault,
		StrikeType: combat.StrikeTypeDefault,
		Element:    attributes.Anemo,
		Durability: 50,
		Mult:       burst[c.TalentLvlBurst()],
	}
	snap := c.Snapshot(&ai)

	// initial hit at 40f
	c.Core.QueueAttackWithSnap(ai, snap, combat.NewCircleHit(c.Core.Combat.Player(), 5, false, combat.TargettableEnemy), 40)

	// field status
	c.Core.Status.Add("jean-q", 600+burstStart)

	// handle user specified amount of In/Out damage
	// TODO: make this work with movement?
	ai.Abil = "Dandelion Breeze (In/Out)"
	ai.Mult = burstEnter[c.TalentLvlBurst()]
	// first enter is at frame 55
	for i := 0; i < enter; i++ {
		c.Core.QueueAttackWithSnap(ai, snap, combat.NewCircleHit(c.Core.Combat.Player(), 5, false, combat.TargettableEnemy), 55+i*delay)
	}

	// handle In/Out damage on field expiry
	c.Core.QueueAttackWithSnap(ai, snap, combat.NewCircleHit(c.Core.Combat.Player(), 5, false, combat.TargettableEnemy), 600+burstStart)

	//heal on cast
	hpplus := snap.Stats[attributes.Heal]
	atk := snap.BaseAtk*(1+snap.Stats[attributes.ATKP]) + snap.Stats[attributes.ATK]
	heal := burstInitialHealFlat[c.TalentLvlBurst()] + atk*burstInitialHealPer[c.TalentLvlBurst()]
	healDot := burstDotHealFlat[c.TalentLvlBurst()] + atk*burstDotHealPer[c.TalentLvlBurst()]

	c.Core.Tasks.Add(func() {
		c.Core.Player.Heal(player.HealInfo{
			Caller:  c.Index,
			Target:  -1,
			Message: "Dandelion Breeze",
			Src:     heal,
			Bonus:   hpplus,
		})
	}, burstStart)

	self, ok := c.Core.Combat.Player().(*avatar.Player)
	if !ok {
		panic("target 0 should be Player but is not!!")
	}

	//attack self
	selfSwirl := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Dandelion Breeze (Self Swirl)",
		Element:    attributes.Anemo,
		Durability: 25,
	}

	// duration is ~10.6s, first tick starts at frame 100, + 60 each
	for i := 100; i <= 600+burstStart; i += 60 {
		c.Core.Tasks.Add(func() {
			// heal
			// c.Core.Log.NewEvent("jean q healing", glog.LogCharacterEvent, c.Index, "+heal", hpplus, "atk", atk, "heal amount", healDot)
			c.Core.Player.Heal(player.HealInfo{
				Caller:  c.Index,
				Target:  c.Core.Player.Active(),
				Message: "Dandelion Field",
				Src:     healDot,
				Bonus:   hpplus,
			})

			// self swirl
			ae := combat.AttackEvent{
				Info:        selfSwirl,
				Pattern:     combat.NewDefSingleTarget(0, combat.TargettablePlayer),
				SourceFrame: c.Core.F,
			}
			c.Core.Log.NewEvent("jean self swirling", glog.LogCharacterEvent, c.Index)
			self.ReactWithSelf(&ae)

			// C4
			if c.Base.Cons >= 4 {
				c.c4()
			}
		}, i)
	}

	c.SetCDWithDelay(action.ActionBurst, 1200, 38)
	// handle energy delay and a4
	c.Core.Tasks.Add(func() {
		c.Energy = 16 //jean a4
	}, 41)

	return action.ActionInfo{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSwap], // earliest cancel
		State:           action.BurstState,
	}
}
