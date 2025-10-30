module duck/apps/namespace1/app1

go 1.23

require (
	duck/common v0.0.0
	duck/httputils v0.0.0
)

replace (
	duck/common => ../../../packages/go/common
	duck/httputils => ../../../packages/go/httputils
)
