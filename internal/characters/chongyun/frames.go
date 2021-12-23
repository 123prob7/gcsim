package chongyun

import "github.com/genshinsim/gcsim/pkg/core"

func (c *char) ActionFrames(a core.ActionType, p map[string]int) (int, int) {
	switch a {
	case core.ActionAttack:
		f := 0
		switch c.NormalCounter {
		//TODO: need to add atkspd mod
		case 0:
			f = 24 //frames from keqing lib
		case 1:
			f = 62 - 24
		case 2:
			f = 124 - 62
		case 3:
			f = 204 - 124
		}
		f = int(float64(f) / (1 + c.Stats[core.AtkSpd]))
		return f, f
	case core.ActionCharge:
		return 30, 30 //frames from keqing lib
	case core.ActionSkill:
		return 57, 57
	case core.ActionBurst:
		return 135, 135 //ok
	default:
		c.Core.Log.Warnf("%v: unknown action (%v), frames invalid", c.Base.Key.String(), a)
		return 0, 0
	}
}