package env

type Environment interface {
	HasFlag(flag string) bool
}

type StaticEnvironment struct {
	Flags []string
}

func (e *StaticEnvironment) HasFlag(flag string) bool {
	for _, f := range e.Flags {
		if flag == f {
			return true
		}
	}
	return false
}

func GetEnvironment() Environment {
	return &StaticEnvironment{
		Flags: getFlags(),
	}
}

func getFlags() []string {
	flags := []string{}
	flags = append(flags, getLinuxFlags()...)
	flags = append(flags, getAzureFlags()...)
	flags = append(flags, getK8sFlags()...)

	return flags
}
