package batch

type StaticBatchDiscoverer struct {
	Machines []string
}

func (d *StaticBatchDiscoverer) Discover() ([]string, error) {
	return d.Machines, nil
}
