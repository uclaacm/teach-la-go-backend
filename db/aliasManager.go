package db

import (
	"context"
	"strings"
	"strconv"
	"errors"
	"fmt"
	"math/rand"
	tinycrypt "github.com/uclaacm/teach-la-go-backend-tinycrypt" 
	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

type Shard struct {
	Count int64
}

var crypt = tinycrypt.Encrypter{
	[]uint64 {
		0x33684192D,
		0x28DAB6A5A,
		0xA928A9246,
		0x224D23A42,
	},
}




func (d *DB) InitShards(ctx context.Context, path string)(error){

	// get the document that has the counters
	doc := d.Collection(path).Doc("--shards--")

	// get the subcollection with the sub-counters
	col := doc.Collection("shards")

	// Initialize each to 0
	for i := 0; i < num_shards; i++{

		d :=  col.Doc(strconv.Itoa(i))
		fmt.Print(d.ID)
		_, err := col.Doc(strconv.Itoa(i)).Set(ctx, map[string]interface{}{
			"Count" : 0,
		})
		
		if err != nil {
			return err
		}
	} 

	// en.InitializeEncrypter(keys)

	return nil

}

func (d *DB) GetID(ctx context.Context, path string)(int64, error){

	// get the document that has the counters
	doc := d.Collection(path).Doc("--shards--")

	// get every document in the collection
	shards := doc.Collection("shards").Documents(ctx)
	count := make([]int64, num_shards)

	// check each shard 
	i := 0
	for {

		d, e := shards.Next()
		if e == iterator.Done {
			break
		}

		if e != nil {
			return 0, errors.New("Fail: high traffic")
		}

		// get how many aliases have been generated, 
		// and divide it by 1024. Record the result
		count[i] = ( d.Data()["Count"].(int64) ) / 1024
		i++
	}
	
	

	for j := range count {
		count[j] = max_size - count[j] - 1
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
	total += 1
	// select a random number
	shard_select := rand.Int63n(total)
	shard_id := int64(0)

	// based on the random number, determine which shard we should choose
	for j := range count {
		shard_select -= count[j]

		//if select_shard is <= to zero, this is the shard we want to select
		if shard_select <= 0 {
			shard_id = int64(j) 
			break 
		}
	}

	// get the subcollection with the sub-counters
	col := doc.Collection("shards")

	shard_ref := col.Doc(strconv.FormatInt(shard_id, 10))

	uid := int64(0)

	// try func until success, in which case nil will be returned
	// Then, the change will be comitted. Otherwise, RunTransaction.. will retry 
	// In Go, this function is blocking,
	// On success, a nil will be returned. 
	err := d.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error{

		doc, err := tx.Get(shard_ref)

		if err != nil {
			return err
		}

		tid, err := doc.DataAt("Count")

		uid = tid.(int64)

		if uid >= shard_cap {
			return errors.New("shard full")
		}

		err = tx.Set(shard_ref, map[string]interface{}{
			"Count": uid + 1,
		}, firestore.MergeAll)

		if err != nil {
			return err
		}

		return err
	})

	if err != nil {
		return 0, errors.New("Fail: high traffic")
	}

	uid += slot_per_shard * shard_id

	
	// keys :=	[]uint64 {
	// 		0x33684192D,
	// 		0x28DAB6A5A,
	// 		0xA928A9246,
	// 		0x224D23A42,
	// 	}
	
	// var crypt tinycrypt.Encrypter
	
	//crypt.InitializeEncrypter(keys)

	wid := int64(crypt.Encrypt24(uint64(uid)))

	return wid, nil

}


// MakeAlias takes an id (usually pid or cid), generates a 3 word id(wid), and 
// stores it in Firebase. The generated wid is returned as a string, with words comma seperated
func (d *DB) MakeAlias(ctx context.Context, uid string, path string) (string, error) {

	// convert uid into a 36 bit hash
	//aid := tinycrypt.MakeHash(uid) 
	//aid := tinycrypt.GenerateHash() 
	aid, err := d.GetID(ctx, path)

	// convert that to a 3 word id
	wid_list := tinycrypt.GenerateWord24(uint64(aid))
	// the result is an array,so concat into a single string
	wid := strings.Join(wid_list, ",") 

	// // get the mapping collection
	 col := d.Collection(path)
	// // get the snapshot of the document with the requested wid
	// snap, err := col.Doc(wid).Get(ctx)

	// //if the doc id is taken, generate a different wid
	// for snap.Exists() == true {

	// 	aid++
	// 	if aid >= 0xFFFFFFFFF{
	// 		aid = 0
	// 	}

	// 	wid_list = tinycrypt.GenerateWord24(aid)
	// 	wid = strings.Join(wid_list, ",") 
		
	// 	snap, err = col.Doc(wid).Get(ctx)
	// }
	
	//create mapping
	_, err = col.Doc(wid).Set(ctx, map[string]interface{}{
		"target" : uid,
	})

	return strings.Join(wid_list, ","), err

}



// GetUIDFromWID returns the UID given a WID
func (d *DB) GetUIDFromWID(ctx context.Context, wid string, path string) (string, error) {

	// get the document with the mapping
	doc, err := d.Collection(path).Doc(wid).Get(ctx)
	if err != nil {
		return "", err
	}

	t := struct {
		Target	string `firestore:target`
	}{}

	err = doc.DataTo(&t)

	return t.Target, err
}
