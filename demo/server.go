package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/go-distributed/epaxos/message"
	"github.com/go-distributed/epaxos/replica"
	"github.com/go-distributed/epaxos/transporter"
	"github.com/golang/glog"
)

var _ = fmt.Printf

const (
	chars           = "ABCDEFG"
	prepareInterval = 1 // 1 seconds
)

type Voter struct {
	r *replica.Replica
}

func (v *Voter) SetReplica(r *replica.Replica) {
	v.r = r
}

// NOTE: This is not idempotent.
//
//	Same command might be executed for multiple times
//	but the exection is slow now, so it is unlikely to happen
func (v *Voter) Execute(c []message.Command) ([]interface{}, error) {
	rate := 0.0
	if v.r != nil {
		rate = v.r.ConflictRate()
	}
	if c == nil || len(c) == 0 {
		fmt.Fprintf(os.Stderr, "From: No op | conflict rate: %.4f\n", rate)
	} else {
		for i := range c {
			fmt.Fprintf(os.Stderr, "%s | conflict rate: %.4f\n", string(c[i]), rate)
		}
	}
	return nil, nil
}

func (v *Voter) HaveConflicts(c1 []message.Command, c2 []message.Command) bool {
	return true
}

func main() {
	var id int
	var restore bool
	var concurrent int

	flag.IntVar(&id, "id", -1, "id of the server")
	flag.BoolVar(&restore, "restore", false, "if recover")
	flag.IntVar(&concurrent, "concurrent", 1, "number of concurrent proposers (1 = sequential mode)")

	flag.Parse()

	if id < 0 {
		fmt.Println("id is required!")
		flag.PrintDefaults()
		return
	}

	addrs := []string{
		":9000", ":9001", ":9002",
		//":9003", ":9004",
	}

	tr, err := transporter.NewUDPTransporter(addrs, uint8(id), len(addrs))
	if err != nil {
		panic(err)
	}
	voter := &Voter{}
	param := &replica.Param{
		Addrs:            addrs,
		ReplicaId:        uint8(id),
		Size:             uint8(len(addrs)),
		StateMachine:     voter,
		Transporter:      tr,
		EnablePersistent: true,
		Restore:          restore,
		TimeoutInterval:  time.Second,
		//ExecuteInterval:  time.Second,
	}
	if restore {
		fmt.Fprintln(os.Stderr, "===restore===")
	}

	fmt.Println("====== Spawn new replica ======")
	r, err := replica.New(param)
	if err != nil {
		glog.Fatal(err)
	}
	voter.SetReplica(r)

	fmt.Println("Done!")
	fmt.Printf("Wait %d seconds to start\n", prepareInterval)
	time.Sleep(prepareInterval * time.Second)
	err = r.Start()
	if err != nil {
		glog.Fatal(err)
	}
	fmt.Println("====== start ======")

	rand.Seed(time.Now().UTC().UnixNano())

	numProposers := concurrent
	if numProposers < 1 {
		numProposers = 1
	}

	fmt.Printf("Running with %d concurrent proposer(s)\n", numProposers)

	var wg sync.WaitGroup
	wg.Add(numProposers)

	for p := 0; p < numProposers; p++ {
		proposerId := p
		go func() {
			defer wg.Done()
			counter := proposerId * 10000 // offset counters to avoid confusion
			for i := 0; i < 1000; i++ {
				time.Sleep(time.Millisecond * 50)
				c := "From: " + strconv.Itoa(id) + ", Proposer: " + strconv.Itoa(proposerId) + ", Command: " + strconv.Itoa(id) + ":" + strconv.Itoa(counter) + ", " + time.Now().String()
				counter++

				cmds := make([]message.Command, 0)
				cmds = append(cmds, message.Command(c))
				r.Propose(cmds...)
			}
		}()
	}

	// Wait for all proposers to finish
	wg.Wait()

	// Give some time for all instances to execute
	time.Sleep(5 * time.Second)

	// Print statistics
	total := r.GetTotalProposals()
	fast := r.GetFastPathCount()
	slow := r.GetSlowPathCount()
	fastPercent := 0.0
	slowPercent := 0.0

	if total > 0 {
		fastPercent = float64(fast) * 100.0 / float64(total)
		slowPercent = float64(slow) * 100.0 / float64(total)
	}

	fmt.Println("\n========== STATISTICS ==========")
	fmt.Printf("Replica ID: %d\n", id)
	fmt.Printf("Total Proposals: %d\n", total)
	fmt.Printf("Fast Path: %d (%.2f%%)\n", fast, fastPercent)
	fmt.Printf("Slow Path: %d (%.2f%%)\n", slow, slowPercent)
	fmt.Printf("Conflict Rate: %.4f\n", r.ConflictRate())
	fmt.Println("================================\n")
}
