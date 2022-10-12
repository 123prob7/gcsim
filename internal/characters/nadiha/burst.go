package nadiha

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/glog"
)

var burstFrames []int

const burstKey = "nadihaburst"

// TODO: check frame
func init() {
	burstFrames = frames.InitAbilSlice(128) // Q -> D/J
	burstFrames[action.ActionAttack] = 127  // Q -> N1
	burstFrames[action.ActionCharge] = 127  // Q -> CA
	burstFrames[action.ActionSkill] = 127   // Q -> E
	burstFrames[action.ActionWalk] = 127    // Q -> W
	burstFrames[action.ActionSwap] = 128    // Q -> Swap
}

func (c *char) Burst(p map[string]int) action.ActionInfo {
	if c.Base.Cons >= 6 {
		c.c6Stacks = 6
		c.AddStatus(nadihaC6Key, 10*60, false)
	}

	hydroBonusDuration := int(burstHydroBonus[c.hydroCount][c.TalentLvlBurst()] * 60)
	c.Core.Status.Add(burstKey, 15*60+hydroBonusDuration)

	// c.a1Src = c.Core.F
	// c.Core.Tasks.Add(c.a1QueueUpdateEMTask(c.a1Src), 1)

	// Cannot be prefed particles
	// TODO: check frame
	c.ConsumeEnergy(10)
	c.SetCDWithDelay(action.ActionBurst, 13.5*60, 10)

	c.Core.Log.NewEvent("nadihaburst added", glog.LogCharacterEvent, c.Index).
		Write("expiry", 15*60+hydroBonusDuration).
		Write("hydroBonusDuration", hydroBonusDuration)

	return action.ActionInfo{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSwap],
		State:           action.BurstState,
	}
}
