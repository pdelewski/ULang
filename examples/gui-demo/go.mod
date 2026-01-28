module gui-demo

go 1.24

require (
	libs/gui v0.0.0
	runtime/graphics v0.0.0
)

replace runtime/graphics => ../../runtime/graphics

replace libs/gui => ../../libs/gui
