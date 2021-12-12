package env

type Environment struct {
	flags []string
}

func (e *Environment) HasFlag(flag string) bool {
	for _, f := range e.flags {
		if flag == f {
			return true
		}
	}
	return false
}

func GetEnvironment() *Environment {
	return &Environment{
		flags: getFlags(),
	}
}

func getFlags() []string {
	flags := []string{}
	flags = append(flags, getLinuxFlags()...)
	flags = append(flags, getAzureFlags()...)
	return flags
}
