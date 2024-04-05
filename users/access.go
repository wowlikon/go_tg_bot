package users

type Access int

const (
	Unregistered Access = iota // EnumIndex = 0
	Waiting                    // EnumIndex = 1
	Member                     // EnumIndex = 2
	Admin                      // EnumIndex = 3
	SU                         // EnumIndex = 4
)

func (w Access) String() string {
	return [...]string{"Unregistered", "Waiting", "Member", "Admin", "SU"}[w]
}

func (w Access) EnumIndex() int {
	return int(w)
}

func AccessList() []Access {
	return []Access{0, 1, 2, 3, 4}
}
