package tftest

default allowed = false

allowed {
	input.planned_values.root_module.resources[_].values.ports[_].external > 10000
}
