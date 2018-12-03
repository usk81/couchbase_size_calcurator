package commands

import (
	"fmt"

	humanize "github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
	"gopkg.in/AlecAivazis/survey.v1"
)

type request struct {
	NumOfDocs            int    `survey:"documents_num"`
	IDSize               int    `survey:"ID_size"`
	ValueSize            int    `survey:"value_size"`
	NumberOfReplicas     int    `survey:"number_of_replicas"`
	WorkingSetPercentage int    `survey:"working_set_percentage"`
	CouchbaseVersion     string `survey:"couchbase_version"`
	DiskType             string `survey:"disk_type"`
}

type sizingDataServiceNodes struct {
	NumberOfCopies  int
	TotalMetadata   int
	TotalDataset    int
	WorkingSet      int
	ClusterRAMQuota int
	NumberOfNodes   int
}

const (
	highWaterMark          = 0.85
	metadataVer2           = 64
	metadata               = 56
	headroomSSD            = 0.25
	headroomHDD            = 0.3
	couchbaseVersionLower  = "2.0.x"
	couchbaseVersionHigher = "2.1 and higher"
	diskTypeHDD            = "HDD"
	diskTypeSSD            = "SSD"
)

var (
	dataCmd = &cobra.Command{
		Use: "data",
		Run: dataCommand,
	}

	qs = []*survey.Question{
		{
			Name: "documents_num",
			Prompt: &survey.Input{
				Message: "How many documents will you save? (e.g. 1000000",
			},
			Validate: survey.Required,
		},
		{
			Name: "ID_size",
			Prompt: &survey.Input{
				Message: "What is the ID size per data? (byte) (e.g. 100",
			},
			Validate: survey.Required,
		},
		{
			Name: "value_size",
			Prompt: &survey.Input{
				Message: "What is the average size per document? (byte) (e.g. 10000",
			},
			Validate: survey.Required,
		},
		{
			Name: "number_of_replicas",
			Prompt: &survey.Input{
				Message: "How much number of replicas? (e.g. 1",
			},
			Validate: survey.Required,
		},
		{
			Name: "working_set_percentage",
			Prompt: &survey.Input{
				Message: "What percentage will you use for working set? (e.g. 20",
			},
			Validate: survey.Required,
		},
		{
			Name: "couchbase_version",
			Prompt: &survey.Select{
				Message: "Which is your couchbase version?",
				Options: []string{couchbaseVersionLower, couchbaseVersionHigher},
				Default: couchbaseVersionHigher,
			},
			Validate: survey.Required,
		},
		{
			Name: "disk_type",
			Prompt: &survey.Select{
				Message: "Which is your disk type?",
				Options: []string{diskTypeHDD, diskTypeSSD},
				Default: diskTypeSSD,
			},
			Validate: survey.Required,
		},
	}
)

func dataCommand(cmd *cobra.Command, args []string) {
	if err := dataAction(); err != nil {
		Exit(err, 1)
	}
}

func dataAction() (err error) {
	var answers request
	err = survey.Ask(qs, &answers)
	if err != nil {
		return
	}

	mpd := metadata
	if answers.CouchbaseVersion == couchbaseVersionLower {
		mpd = metadataVer2
	}

	var headroom float64
	if answers.DiskType == diskTypeSSD {
		headroom = headroomSSD
	} else {
		headroom = headroomHDD
	}

	// no_of_copies = 1 + number_of_replicas
	nc := 1 + answers.NumberOfReplicas
	//total_metadata = (documents_num) * (metadata_per_document + ID_size) * (no_of_copies)
	tm := answers.NumOfDocs * (mpd + answers.IDSize) * nc
	// total_dataset = (documents_num) * (value_size) * (no_of_copies)
	td := answers.NumOfDocs * answers.ValueSize * nc
	// working_set = total_dataset * (working_set_percentage)
	ws := td * answers.WorkingSetPercentage / 100
	// Cluster RAM quota required = (total_metadata + working_set) * (1 + headroom) / (high_water_mark)
	cmq := int(float64(tm+ws) * (1.0 + headroom) / highWaterMark)

	DisplayResult(sizingDataServiceNodes{
		NumberOfCopies:  nc,
		TotalMetadata:   tm,
		TotalDataset:    td,
		WorkingSet:      ws,
		ClusterRAMQuota: cmq,
	})
	return
}

// DisplayResult displays calcurating result
func DisplayResult(r sizingDataServiceNodes) {
	fmt.Printf("no_of_copies:%d\ntotal_metadata:%s(%d)\ntotal_dataset:%s(%d)\nworking_set:%s(%d)\nCluster RAM quota required:%s(%d)\nnumber of nodes:%d",
		r.NumberOfCopies,
		humanize.Bytes(uint64(r.TotalMetadata)),
		r.TotalMetadata,
		humanize.Bytes(uint64(r.TotalDataset)),
		r.TotalDataset,
		humanize.Bytes(uint64(r.WorkingSet)),
		r.WorkingSet,
		humanize.Bytes(uint64(r.ClusterRAMQuota)),
		r.ClusterRAMQuota,
		r.NumberOfNodes,
	)
}

func init() {
	RootCmd.AddCommand(dataCmd)
}
