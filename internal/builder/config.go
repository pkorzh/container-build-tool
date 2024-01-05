package builder

func (b *Builder) SetWorkingDir(w string) {
	b.OCIImage.Config.WorkingDir = w
}

func (b *Builder) SetEntrypoint(e []string) {
	b.OCIImage.Config.Entrypoint = e
}

func (b *Builder) SetCmd(c []string) {
	b.OCIImage.Config.Cmd = c
}

func (b *Builder) SetUser(u string) {
	b.OCIImage.Config.User = u
}

func (b *Builder) SetPorts(p []string) {
	b.OCIImage.Config.ExposedPorts = make(map[string]struct{})
	for _, port := range p {
		b.OCIImage.Config.ExposedPorts[port] = struct{}{}
	}
}

func (b *Builder) SetOS(os string) {
	b.OCIImage.OS = os
}

func (b *Builder) SetArch(arch string) {
	b.OCIImage.Architecture = arch
}
