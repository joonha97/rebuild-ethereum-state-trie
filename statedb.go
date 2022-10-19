// RebuildStorageTrieFromKeyValue rebuilds a storage trie with hard-coded trie (joonha)
func (s *StateDB) RebuildStorageTrieFromKeyValue(leveldbPath_in string, trieRoot common.Hash, leveldbPath_out_normal string, leveldbPath_out_stack string) { // state trie

	/* config */
	const (
		normalTrie = 1
		stackTrie  = 2
	)
	simulatingTrie := normalTrie

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
		normalTr, _ := trie.New(common.Hash{}, trie.NewDatabase(diskdb_out_normal))
		
		// inject the data
		RebuildNormalTrieStarts := time.Now().UnixNano() / int64(time.Millisecond)
		TrieUpdateStarts := time.Now().UnixNano() / int64(time.Millisecond)
		for i, enc := range accounts {
			data := new(types.StateAccount)

			if err := rlp.DecodeBytes(enc, data); err != nil {
				log.Error("Failed to decode state object", "err", err)
			}

			acc, _ := rlp.EncodeToBytes(data)
			if err := normalTr.TryUpdate(keys[i][:], acc); err != nil {
				s.setError(fmt.Errorf("updateStateObject (%x) error: %v", err))
			}
		}
		TrieUpdateEnds := time.Now().UnixNano() / int64(time.Millisecond)

		// trie commit
		var root common.Hash
		trieCommitStarts := time.Now().UnixNano() / int64(time.Millisecond)
		if root, _, err = normalTr.Commit(nil); err != nil {
			fmt.Println("normalTr.Commit error !!")
		}
		trieCommitEnds := time.Now().UnixNano() / int64(time.Millisecond)

		// db commit
		dbCommitStarts := time.Now().UnixNano() / int64(time.Millisecond)
		normalTr.Database().Commit(root, false, nil)
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
		
		// inject the data and commit
		RebuildStackTrieStarts := time.Now().UnixNano() / int64(time.Millisecond)
		for index, key := range keys {
			stackTr.TryUpdate(key[:], accounts[index])
		}
		if _, err := stackTr.Commit(); err != nil {
			log.Error("Failed to commit stack slots", "err", err)
		}

		RebuildStackTrieEnds := time.Now().UnixNano() / int64(time.Millisecond)
		fmt.Println("$$$ rebuilding the stack trie done")
		fmt.Println("Rebuild Stack Trie Time Duration: ", RebuildStackTrieEnds-RebuildStackTrieStarts, "(ms)")
		fmt.Println("stackTr.Hash(): ", stackTr.Hash())
	}
}
