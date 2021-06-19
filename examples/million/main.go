// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/kelindar/column"
)

func main() {
	amount := 1000000
	players := column.NewCollection(column.Options{
		Capacity: amount,
	})

	// insert the data first
	measure("insert", fmt.Sprintf("%v rows", amount), func() {
		createCollection(players, amount)
	})

	// run a full scan
	measure("full scan", "age >= 30", func() {
		players.Query(func(txn *column.Txn) error {
			count := txn.WithFloat("age", func(v float64) bool {
				return v >= 30
			}).Count()
			println("-> result =", count)
			return nil
		})
	})

	// run a query over human mages
	measure("indexed query", "human mages", func() {
		players.Query(func(txn *column.Txn) error {
			println("-> result =", txn.With("human", "mage").Count())
			return nil
		})
	})

	// update everyone
	measure("update", "balance of everyone", func() {
		updates := 0
		players.Query(func(txn *column.Txn) error {
			return txn.Range("balance", func(v column.Cursor) bool {
				updates++
				v.Update(1000.0)
				return true
			})
		})
		fmt.Printf("-> updated %v rows\n", updates)
	})

	// update age of mages
	measure("update", "age of mages", func() {
		updates := 0
		players.Query(func(txn *column.Txn) error {
			return txn.With("mage").Range("age", func(v column.Cursor) bool {
				updates++
				v.Update(99.0)
				return true
			})
		})
		fmt.Printf("-> updated %v rows\n", updates)
	})
}

// createCollection loads a collection of players
func createCollection(out *column.Collection, amount int) *column.Collection {
	out.CreateColumn("serial", column.ForAny())
	out.CreateColumn("name", column.ForAny())
	out.CreateColumn("active", column.ForBool())
	out.CreateColumn("class", column.ForEnum())
	out.CreateColumn("race", column.ForEnum())
	out.CreateColumn("age", column.ForFloat64())
	out.CreateColumn("hp", column.ForFloat64())
	out.CreateColumn("mp", column.ForFloat64())
	out.CreateColumn("balance", column.ForFloat64())
	out.CreateColumn("gender", column.ForEnum())
	out.CreateColumn("guild", column.ForEnum())
	out.CreateColumn("location", column.ForAny())

	// index for humans
	out.CreateIndex("human", "race", func(v interface{}) bool {
		return v == "human"
	})

	// index for mages
	out.CreateIndex("mage", "class", func(v interface{}) bool {
		return v == "mage"
	})

	// Load the 500 rows from JSON
	b, err := os.ReadFile("../../fixtures/players.json")
	if err != nil {
		panic(err)
	}

	// Unmarshal the items
	var data []map[string]interface{}
	if err := json.Unmarshal(b, &data); err != nil {
		panic(err)
	}

	// Load and copy until we reach the amount required
	for i := 0; i < amount/len(data); i++ {
		if i%200 == 0 {
			fmt.Printf("-> inserted %v rows\n", out.Count())
		}

		out.Query(func(txn *column.Txn) error {
			for _, p := range data {
				txn.Insert(p)
			}
			return nil
		})
	}
	return out
}

// measure runs a function and measures it
func measure(action, name string, fn func()) {
	defer func(start time.Time) {
		fmt.Printf("-> %v took %v\n", action, time.Since(start).String())
	}(time.Now())

	fmt.Println()
	fmt.Printf("running %v of %v...\n", action, name)
	fn()
}