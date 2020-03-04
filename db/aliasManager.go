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
)

type Shard struct {
	Count int64
}

var num_shards = 8 

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

	return nil

}

func (d *DB) GetID(ctx context.Context, path string)(int64, error){

	// get the document that has the counters
	doc := d.Collection(path).Doc("--shards--")

	// get the subcollection with the sub-counters
	col := doc.Collection("shards")

	doc_id := strconv.Itoa(rand.Intn(num_shards))

	shard_ref := col.Doc(doc_id)

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

	return uid, nil

}


// MakeAlias takes an id (usually pid or cid), generates a 3 word id(wid), and 
// stores it in Firebase. The generated wid is returned as a string, with words comma seperated
func (d *DB) MakeAlias(ctx context.Context, uid string, path string) (string, error) {

	// convert uid into a 36 bit hash
	//aid := tinycrypt.MakeHash(uid) 
	aid := tinycrypt.GenerateHash() 

	// convert that to a 3 word id
	wid_list := tinycrypt.GenerateWord36(aid)
	// the result is an array,so concat into a single string
	wid := strings.Join(wid_list, ",") 

	// get the mapping collection
	col := d.Collection(path)
	// get the snapshot of the document with the requested wid
	snap, err := col.Doc(wid).Get(ctx)

	//if the doc id is taken, generate a different wid
	for snap.Exists() == true {

		aid++
		if aid >= 0xFFFFFFFFF{
			aid = 0
		}

		wid_list = tinycrypt.GenerateWord36(aid)
		wid = strings.Join(wid_list, ",") 
		
		snap, err = col.Doc(wid).Get(ctx)
	}
	
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
