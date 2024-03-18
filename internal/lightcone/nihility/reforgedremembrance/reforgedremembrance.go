package reforgedremembrance

import (
	"github.com/simimpact/srsim/pkg/engine"
	"github.com/simimpact/srsim/pkg/engine/equip/lightcone"
	"github.com/simimpact/srsim/pkg/engine/event"
	"github.com/simimpact/srsim/pkg/engine/info"
	"github.com/simimpact/srsim/pkg/engine/modifier"
	"github.com/simimpact/srsim/pkg/engine/prop"
	"github.com/simimpact/srsim/pkg/key"
	"github.com/simimpact/srsim/pkg/model"
)

const (
	rememberance key.Modifier = "reforged-rememberance" // rememberance = Prophet stack if i do this correctly
	atkBuff      key.Modifier = "reforged-rememberance-atk-buff"
	defShred     key.Modifier = "reforged-rememberance-def-shred"
)

type state struct {
	atkBuff, defShred float64
}

// Increases the wearer's Effect Hit Rate by 40%. When the wearer deals DMG to an enemy
// inflicted with Wind Shear, Burn, Shock, or Bleed, each respectively grants 1 stack of Prophet,
// stacking up to 4 time(s). In a single battle, only 1 stack of Prophet can be granted for each
// type of DoT. Every stack of Prophet increases wearer's ATK by 5% and enables the DoT dealt
// to ignore 7.2% of the target's DEF.
func init() {
	lightcone.Register(key.ReforgedRemembrance, lightcone.Config{
		CreatePassive: Create,
		Rarity:        5,
		Path:          model.Path_NIHILITY,
		Promotions:    promotions,
	})
	modifier.Register(rememberance, modifier.Config{
		Listeners: modifier.Listeners{
			OnAfterHit: addProphetStack,
		},
	})
	modifier.Register(atkBuff, modifier.Config{
		Stacking:          modifier.ReplaceBySource,
		StatusType:        model.StatusType_STATUS_BUFF,
		MaxCount:          4,
		CountAddWhenStack: 1,
		Listeners: modifier.Listeners{
			OnAdd: recalcAtkBuff,
		},
	})
	modifier.Register(defShred, modifier.Config{
		BehaviorFlags: []model.BehaviorFlag{
			model.BehaviorFlag_STAT_DEF_DOWN,
		},
		Stacking:          modifier.ReplaceBySource,
		StatusType:        model.StatusType_STATUS_BUFF,
		MaxCount:          4,
		CountAddWhenStack: 1,
		Listeners: modifier.Listeners{
			OnAdd: recalcDefShred,
		},
	})
}

func Create(engine engine.Engine, owner key.TargetID, lc info.LightCone) {
	ehrAmt := 0.4 + 0.05*float64(lc.Imposition)
	modState := state{
		atkBuff:  0.05 + 0.01*float64(lc.Imposition),
		defShred: 0.072 + 0.07*float64(lc.Imposition),
	}
	engine.AddModifier(owner, info.Modifier{
		Name:   rememberance,
		Source: owner,
		Stats:  info.PropMap{prop.EffectHitRate: ehrAmt},
		State:  &modState,
	})
}

func addProphetStack(mod *modifier.Instance, e event.HitEnd) {
	state := mod.State().(*state)
	if mod.Engine().HasBehaviorFlag(e.Defender, model.BehaviorFlag_STAT_DOT_ELECTRIC, model.BehaviorFlag_STAT_DOT_BURN, model.BehaviorFlag_STAT_DOT_BLEED) {
		mod.Engine().AddModifier(mod.Owner(), info.Modifier{
			Name:   atkBuff,
			Source: mod.Owner(),
			State:  state.atkBuff,
		})
		mod.Engine().AddModifier(e.Defender, info.Modifier{
			Name:   defShred,
			Source: mod.Owner(),
			State:  state.defShred,
		})
	}
}

func recalcAtkBuff(mod *modifier.Instance) {
	atkBuff := mod.State().(float64) * mod.Count()
	mod.AddProperty(prop.ATKPercent, atkBuff)
}

func recalcDefShred(mod *modifier.Instance) {
	defShred := mod.State().(float64) * mod.Count()
	mod.AddProperty(prop.DEFPercent, defShred)
}
