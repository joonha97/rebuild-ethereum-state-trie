// RebuildStorageTrieFromKeyValue rebuilds a trie with hard-coded trie (joonha)
func (s *StateDB) RebuildStorageTrieFromKeyValue(leveldbPath_in string, trieRoot common.Hash, leveldbPath_out_normal string, leveldbPath_out_stack string) { // state trie

	/* config */
	const (
		normalTrie = 1
		stackTrie  = 2
	)
	simulatingTrie := stackTrie

	/* load key-value */
	common.LoadAccountsAndKeys()
	accounts := common.Accounts
	keys := common.Keys


	/* normal trie */
	if simulatingTrie == normalTrie {
		
		// new trie
		diskdb_out_normal, err := leveldb.New(leveldbPath_out_normal, leveldbCache, leveldbHandles, leveldbNamespace, leveldbReadonly)
		if err != nil {
			fmt.Println("leveldb.New error (2) !! ->", err)
			os.Exit(1)
		}
		rebuiltTrie, _ = trie.NewSecure(common.Hash{}, trie.NewDatabase(diskdb_out_normal))
		
		// inject the data
		RebuildNormalTrieStarts := time.Now().UnixNano() / int64(time.Millisecond)
		TrieUpdateStarts := time.Now().UnixNano() / int64(time.Millisecond)
		for i, enc := range accounts {
			data := new(types.StateAccount)
			addr := common.BytesToAddress(enc) // BytesToAddress() turns last 20 bytes into addr

			if err := rlp.DecodeBytes(enc, data); err != nil {
				log.Error("Failed to decode state object", "addr", addr, "err", err)
			}

			acc, _ := rlp.EncodeToBytes(data)
			if err := rebuiltTrie.TryUpdate_SetKey(keys[i][:], acc); err != nil {
				s.setError(fmt.Errorf("updateStateObject (%x) error: %v", addr[:], err))
			}
		}
		TrieUpdateEnds := time.Now().UnixNano() / int64(time.Millisecond)

		// trie commit
		var root common.Hash
		trieCommitStarts := time.Now().UnixNano() / int64(time.Millisecond)
		if root, _, err = rebuiltTrie.Commit(nil); err != nil {
			fmt.Println("rebuiltTrie.Commit error !!")
		}
		trieCommitEnds := time.Now().UnixNano() / int64(time.Millisecond)

		// db commit
		dbCommitStarts := time.Now().UnixNano() / int64(time.Millisecond)
		rebuiltTrie.Database().Commit(root, false, nil)
		dbCommitEnds := time.Now().UnixNano() / int64(time.Millisecond)

		RebuildNormalTrieEnds := time.Now().UnixNano() / int64(time.Millisecond)

		fmt.Println("$$$ rebuilding the normal trie done")
		fmt.Println("Update Trie Time Duration: ", TrieUpdateEnds-TrieUpdateStarts, "(ms)")
		fmt.Println("Trie Commit Time Duration: ", trieCommitEnds-trieCommitStarts, "(ms)")
		fmt.Println("Db Commit Time Duration: ", dbCommitEnds-dbCommitStarts, "(ms)")
		fmt.Println("Rebuild Normal Trie Time Duration: ", RebuildNormalTrieEnds-RebuildNormalTrieStarts, "(ms)")
	}

	/* stack trie */
	if simulatingTrie == stackTrie {
		
		// new trie
		diskdb_out_stack, err := leveldb.New(leveldbPath_out_stack, leveldbCache, leveldbHandles, leveldbNamespace, leveldbReadonly)
		if err != nil {
			fmt.Println("leveldb.New error (3) !! ->", err)
			os.Exit(1)
		}
		stackTr := trie.NewStackTrie(diskdb_out_stack.NewBatch())
		
		// inject the data
		RebuildStackTrieStarts := time.Now().UnixNano() / int64(time.Millisecond)
		for index, key := range keys {
			stackTr.TryUpdate(key[:], accounts[index])
		}

		
		RebuildStackTrieEnds := time.Now().UnixNano() / int64(time.Millisecond)
		
		fmt.Println("$$$ rebuilding the stack trie done")
		fmt.Println("Rebuild Stack Trie Time Duration: ", RebuildStackTrieEnds-RebuildStackTrieStarts, "(ms)")
		fmt.Println("stackTr.Hash(): ", stackTr.Hash())
	}
}
