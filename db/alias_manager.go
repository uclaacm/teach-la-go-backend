package db

import (
	"context"
	"errors"
	"math/rand"
	"strconv"
	"strings"

	"cloud.google.com/go/firestore"
	tinycrypt "github.com/uclaacm/teach-la-go-backend-tinycrypt"
	"google.golang.org/api/iterator"
)

// Shard class
// Currently all it does is store the number of shards
type Shard struct {
	Count int64
}

// A structure to store the key.
// This is not a secret key of any kind, all it does is scramble the order of ID
// so classes created around the same time will not have similar ID's
var crypt = tinycrypt.Encrypter{
	Key: []uint64{
		0x33684192D,
		0x28DAB6A5A,
		0xA928A9246,
		0x224D23A42,
	},
}

// InitShards takes the path to the class_alias collection (path)
// and initializes the shard
func (d *DB) InitShards(ctx context.Context, path string) error {

	// get the document that has the counters
	doc := d.Collection(path).Doc(shardName)

	// get the subcollection with the sub-counters
	col := doc.Collection("shards")

	// Initialize each to 0
	for i := 0; i < numShards; i++ {

		_, err := col.Doc(strconv.Itoa(i)).Set(ctx, map[string]interface{}{
			"Count": 0,
		})

		if err != nil {
			return err
		}
	}

	return nil

}

// GetID returns the proper ID from the given alias.
func (d *DB) GetID(ctx context.Context, path string) (int64, error) {
	// get the document that has the counters
	doc := d.Collection(path).Doc("--shards--")

	// get every document in the collection
	shards := doc.Collection("shards").Documents(ctx)
	count := make([]int64, numShards)

	// check each shard
	i := 0
	for {

		d, e := shards.Next()
		if e == iterator.Done {
			break
		}

		if e != nil {
			return 0, errors.New("Error: Failed to access shard")
		}

		// Get how many aliases have been generated, and divide it by the divider factor.
		// This will give the number of blocks that are used in the shard.
		// Record the result
		count[i] = (d.Data()["Count"].(int64)) / divider
		i++
	}

	// Update the array
	// Currently we have the number of blocks *used* per shard
	// Change this to the number of blocks *remaining* per shard
	// The 1 is added so each shard allocates 1 block as a "free block"
	// which will never be used. This is done so we have some extra time
	// in an event all the IDs are used up, we can remove the '1' to release those
	// free blocks, and in the meantime think of a way to allocate more IDs.
	for j := range count {
		count[j] = maxSize - count[j] - 1
	}

	// sum up the numbers
	total := int64(0)
	for j := range count {
		total += count[j]
	}

	// if total is 0, all slots are full
	if total == 0 {
		return 0, errors.New("Server full")
	}
	total++
	// select a random number
	shardSelect := rand.Int63n(total)
	shardID := int64(0)

	// Based on the random number, determine which shard we should choose
	// The amount of blocks left in each shard will determine the probability
	// that shard is chosen. This works as a natural load balancer, which ensures
	// shards that are nearly full gets chosen less frequently, while empty shards are more likely to fill up.
	for j := range count {
		shardSelect -= count[j]

		// if select_shard is <= to zero, this is the shard we want to select
		if shardSelect <= 0 {
			shardID = int64(j)
			break
		}
	}

	// get the subcollection with the sub-counters
	col := doc.Collection("shards")

	shardRef := col.Doc(strconv.FormatInt(shardID, 10))

	uid := int64(0)

	// try func until success, in which case nil will be returned
	// Then, the change will be committed. Otherwise, RunTransaction will retry until timeout
	// In Go, this function is blocking,
	err := d.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {

		doc, err := tx.Get(shardRef)

		if err != nil {
			return err
		}

		tid, err := doc.DataAt("Count")
		if err != nil {
			return err
		}

		uid = tid.(int64)

		if uid >= shardCap {
			return errors.New("shard full")
		}

		err = tx.Set(shardRef, map[string]interface{}{
			"Count": uid + 1,
		}, firestore.MergeAll)

		if err != nil {
			return err
		}

		return err
	})

	if err != nil {
		return 0, errors.New("Error: Shard update transaction failed, possible high traffic")
	}

	uid += slotPerShard * shardID

	wid := int64(crypt.Encrypt24(uint64(uid)))

	return wid, nil

}

// MakeAlias takes an id (usually pid or cid), generates a 3 word id(wid), and
// stores it in Firebase. The generated wid is returned as a string, with words comma seperated
func (d *DB) MakeAlias(ctx context.Context, uid string, path string) (string, error) {

	// get a unique ID from the distributed counter
	aid, err := d.GetID(ctx, path)
	if err != nil {
		return "", err
	}

	// convert that to a 2 word id
	widList := tinycrypt.GenerateWord24(uint64(aid))
	// the result is an array,so concat into a single string
	wid := strings.Join(widList, ",")

	// get the mapping collection
	col := d.Collection(path)

	// create mapping between UID and WID
	_, err = col.Doc(wid).Set(ctx, map[string]interface{}{
		"target": uid,
	})

	return strings.Join(widList, ","), err
}

// GetUIDFromWID returns the UID given a WID
func (d *DB) GetUIDFromWID(ctx context.Context, wid string, path string) (string, error) {

	// get the document with the mapping
	doc, err := d.Collection(path).Doc(wid).Get(ctx)
	if err != nil {
		return "", err
	}

	t := struct {
		Target string `firestore:"target"`
	}{}

	err = doc.DataTo(&t)

	return t.Target, err
}
