package users

type Access int

const (
	Unregistered Access = iota // EnumIndex = 0
	NoAccess                   // EnumIndex = 1
	Waiting                    // EnumIndex = 2
	Member                     // EnumIndex = 3
	Admin                      // EnumIndex = 4
	SU                         // EnumIndex = 5
)

func (w Access) String() string {
	return [...]string{"Unregistered", "NoAccess", "Waiting", "Member", "Admin", "SU"}[w]
}

func (w Access) EnumIndex() int {
	return int(w)
}

func AccessList() []Access {
	return []Access{0, 1, 2, 3, 4, 5}
}
