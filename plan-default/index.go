package plan_default

import (
	. "github.com/gobigger/bigger"
)

func Driver() (PlanDriver) {
	return &defaultPlanDriver{}
}


func init() {
	Bigger.Driver("default", Driver())
}

