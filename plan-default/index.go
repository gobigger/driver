package plan_default

import (
	. "github.com/yatlabs/bigger"
)

func Driver() (PlanDriver) {
	return &defaultPlanDriver{}
}


func init() {
	Bigger.Driver("default", Driver())
}

