package main

import (
	"gopkg.in/cheggaaa/pb.v1"
	"os"
)

type Progress struct {
	bar *pb.ProgressBar
}

func (p *Progress) IsEnabled() bool {
	return p.bar != nil
}

func (p *Progress) Enable() {
	p.bar = pb.New(0)
	p.bar.Output = os.Stderr
}

func (p *Progress) Disable() {
	p.bar = nil
}

func (p *Progress) SetTotal(total int) {
	if p.bar != nil {
		p.bar.SetTotal(total)
	}
}

func (p *Progress) Increment() {
	if p.bar != nil {
		p.bar.Increment()
	}
}
