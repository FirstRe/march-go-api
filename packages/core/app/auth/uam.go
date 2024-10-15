package auth

type Action string

const (
	INMaker        Action = "INMaker"
	INChecker      Action = "INChecker"
	SSuperUser     Action = "SSuperUser"
	INViewer       Action = "INViewer"
	INCSV          Action = "INCSV"
	INBranchMaker  Action = "INBranchMaker"
	INBranchViewer Action = "INBranchViewer"
	INBrandMaker   Action = "INBrandMaker"
	INBrandViewer  Action = "INBrandViewer"
	INTypeMaker    Action = "INTypeMaker"
	INTypeViewer   Action = "INTypeViewer"
	INTrashMaker   Action = "INTrashMaker"
)

const (
	Admin      = "ADMIN"
	SuperAdmin = "SUPERADMIN"
	Any        = "ANY"
	Free       = "FREE"
)

var UAMSTRUCT = []string{"AnyAdminScope", "MakerAdminScope", "ViewerAdminScope"}

type UAM struct {
	AnyAdminScope    []Action
	MakerAdminScope  []Action
	ViewerAdminScope []Action
	ScopesMap        map[string][]Action
}

func NewUAM() UAM {
	uam := UAM{
		AnyAdminScope: []Action{
			SSuperUser, INMaker, INViewer, INCSV,
			INBranchMaker, INBranchViewer,
			INBrandMaker, INBrandViewer,
			INTypeMaker, INTypeViewer,
			INTrashMaker,
		},
		MakerAdminScope: []Action{
			INMaker, INCSV, INBranchMaker,
			INBrandMaker, INTypeMaker,
			INTrashMaker,
		},
		ViewerAdminScope: []Action{
			INViewer, INBranchViewer,
			INBrandViewer, INTypeViewer,
		},
	}

	uam.ScopesMap = map[string][]Action{
		"AnyAdminScope":    uam.AnyAdminScope,
		"MakerAdminScope":  uam.MakerAdminScope,
		"ViewerAdminScope": uam.ViewerAdminScope,
	}
	return uam
}
