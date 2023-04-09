package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/mb0/qnpdub/av"
	"github.com/mb0/qnpdub/av/clap"
	"github.com/mb0/qnpdub/av/ffm"
)

func doCat(args []string) error {
	out := args[0]
	o, vs, as := probe(args[1:])
	if len(as) == 0 {
		as = vs
	}
	return o.Concat(out, ffm.Paths(vs), ffm.Paths(as))
}

func doClap(args []string) error {
	o, paths := opts(args)
	d := clap.Default()
	ws, err := d.LoadAll(paths...)
	if err != nil {
		return err
	}
	offs, err := d.Match(o.Fps, ws...)
	if err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(offs)
}

func doSync(args []string) error {
	out := args[0]
	o, vs, as := probe(args[1:])
	if len(as) < 1 || len(vs) < 1 {
		return fmt.Errorf("sync needs at least one video and one audio file")
	}
	// last video and audio index and offset duration
	vl, al := len(vs)-1, len(as)-1
	vlo, alo := sumDur(vs[:vl]), sumDur(as[:al])
	// detect clap in the last video and last audio file
	d := clap.Default()
	ws, err := d.LoadAll(vs[vl].Path, as[al].Path)
	if err != nil {
		return err
	}
	claps, err := d.Match(o.Fps, ws...)
	if err != nil {
		return err
	}
	// get clap in total offset
	vc := vlo + claps[0].Clap
	ac := alo + claps[1].Clap
	// we want to start if we have both video and audio
	// as we usually start recording audio synced to a song
	// TODO calulate offsets respecting opts offs
	if diff := vc - ac; diff >= 0 {
		o.Vod = diff
	} else {
		o.Aod = -diff
	}
	o.Dur = vc - o.Vod
	return o.Concat(out, ffm.Paths(vs), ffm.Paths(as))
}

func sumDur(nfos []*ffm.Info) (sum av.Dur) {
	for _, nfo := range nfos {
		sum += nfo.Format.Dur("duration")
	}
	return sum
}

func opts(args []string) (*ffm.Opts, []string) {
	o := ffm.Def()
	flags := o.Flags()
	err := flags.Parse(args)
	if err != nil {
		log.Fatalf("invalid flag: %v", err)
	}
	return o, flag.Args()
}

func probe(args []string) (_ *ffm.Opts, vs, as []*ffm.Info) {
	o, paths := opts(args)
	nfos, err := o.ProbeAll(paths...)
	if err != nil {
		log.Fatalf("probe failed: %v", err)
	}
	for _, nfo := range nfos {
		if v := nfo.Video(); v != nil {
			vs = append(vs, nfo)
		} else if a := nfo.Audio(); a != nil {
			as = append(as, nfo)
		}
	}
	return o, vs, as
}
