package main

import (
	"github.com/testground/sdk-go/run"
	"github.com/testground/sdk-go/runtime"
)

var testcases = map[string]interface{}{
	"transaction": InitializedTestCaseFn(testTransaction),
	"new_account": InitializedTestCaseFn(testNewAccount),
}

// test_transaction in this test each node starts with getting the ballance
// of the test account. Than the first node transers 100 coins between the
// tests accounts.
// Then all the nodes wait for the next epoch and assert the accounts' balance.
func testTransaction(env *runtime.RunEnv, initCtx *InitContext) {
	t := NewSystemTest(env, initCtx)
	t.Logf("Starting transaction test, %d is up", t.ID)
	// TODO: ready should be replaced with starting the node and waiting for
	// genesis
	t.SetState("ready")
	t.WaitAll("ready")
	start1 := t.GetBalance(Account1)
	start2 := t.GetBalance(Account2)
	if t.ID == 1 {
		t.SendCoins(Account1, Account2, 100)
	}
	t.WaitTillEpoch()
	t.RequireBalance(Account1, start1-100)
	t.RequireBalance(Account2, start2+100)
}

// test_new_account in this test each node starts with getting the ballance
// of the test account. Than the first node creates a new account and
// transers 100 coins between inb from a test account.
// Then all the nodes wait for the next epoch and assert the accounts' balance.

func testNewAccount(env *runtime.RunEnv, initCtx *InitContext) {
	t := NewSystemTest(env, initCtx)
	t.Logf("Starting new account test, %d is up", t.ID)
	// TODO: ready should be replaced with starting the node and waiting for
	// genesis
	t.SetState("ready")
	t.WaitAll("ready")
	start1 := t.GetBalance(Account1)
	if t.ID == 1 {
		account := t.NewAccount()
		t.SendCoins(Account1, account, 100)
	}
	t.WaitTillEpoch()
	t.RequireBalance(Account1, start1-100)
	t.RequireBalance(account, 100)
}

func main() {
	run.InvokeMap(testcases)
}
