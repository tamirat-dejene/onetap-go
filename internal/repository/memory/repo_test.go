package memory_test

import (
	"testing"

	"github.com/onetap/salary-advance/internal/domain"
	"github.com/onetap/salary-advance/internal/repository/memory"
)

func TestCustomerRepo_SeedAndGetAll(t *testing.T) {
	repo := memory.NewCustomerRepo()
	repo.Seed([]*domain.Customer{
		{ID: 1, CustomerName: "ALICE", AccountNo: "ACC001"},
		{ID: 2, CustomerName: "BOB", AccountNo: "ACC002"},
	})

	all := repo.GetAll()
	if len(all) != 2 {
		t.Errorf("expected 2 customers, got %d", len(all))
	}
}

func TestCustomerRepo_GetByAccountNo(t *testing.T) {
	repo := memory.NewCustomerRepo()
	repo.Seed([]*domain.Customer{
		{ID: 1, CustomerName: "ALICE", AccountNo: "ACC001"},
	})

	c, ok := repo.GetByAccountNo("ACC001")
	if !ok {
		t.Fatal("expected to find customer ACC001")
	}
	if c.CustomerName != "ALICE" {
		t.Errorf("expected ALICE, got %q", c.CustomerName)
	}

	_, ok = repo.GetByAccountNo("NONEXISTENT")
	if ok {
		t.Error("expected not found for non-existent account")
	}
}

func TestCustomerRepo_SaveAndGetVerified(t *testing.T) {
	repo := memory.NewCustomerRepo()
	repo.SaveVerified(&domain.Customer{ID: 1, CustomerName: "ALICE", AccountNo: "ACC001"})
	repo.SaveVerified(&domain.Customer{ID: 2, CustomerName: "BOB", AccountNo: "ACC002"})

	verified := repo.GetAllVerified()
	if len(verified) != 2 {
		t.Errorf("expected 2 verified customers, got %d", len(verified))
	}
}

func TestCustomerRepo_VerifiedDoesNotPolluteCandidates(t *testing.T) {
	repo := memory.NewCustomerRepo()
	repo.Seed([]*domain.Customer{{ID: 1, CustomerName: "ALICE", AccountNo: "ACC001"}})
	repo.SaveVerified(&domain.Customer{ID: 99, CustomerName: "EXTRA", AccountNo: "ACC099"})

	all := repo.GetAll()
	if len(all) != 1 {
		t.Errorf("canonical store should have 1, got %d", len(all))
	}
}

func TestTransactionRepo_SeedAndGetAll(t *testing.T) {
	repo := memory.NewTransactionRepo()
	repo.Seed([]*domain.Transaction{
		{ID: "t1", FromAccount: "ACC001", Amount: 100},
		{ID: "t2", FromAccount: "ACC002", Amount: 200},
	})

	all := repo.GetAll()
	if len(all) != 2 {
		t.Errorf("expected 2 transactions, got %d", len(all))
	}
}

func TestTransactionRepo_GetByAccountNo(t *testing.T) {
	repo := memory.NewTransactionRepo()
	repo.Seed([]*domain.Transaction{
		{ID: "t1", FromAccount: "ACC001", Amount: 100},
		{ID: "t2", FromAccount: "ACC001", Amount: 200},
		{ID: "t3", FromAccount: "ACC002", Amount: 300},
	})

	txs := repo.GetByAccountNo("ACC001")
	if len(txs) != 2 {
		t.Errorf("expected 2 txs for ACC001, got %d", len(txs))
	}

	empty := repo.GetByAccountNo("NOBODY")
	if len(empty) != 0 {
		t.Errorf("expected 0 txs, got %d", len(empty))
	}
}

func TestTransactionRepo_Save(t *testing.T) {
	repo := memory.NewTransactionRepo()
	repo.Save(&domain.Transaction{ID: "t1", FromAccount: "ACC001", Amount: 50})

	txs := repo.GetByAccountNo("ACC001")
	if len(txs) != 1 {
		t.Errorf("expected 1 tx, got %d", len(txs))
	}
}

func TestUserRepo_SaveAndGet(t *testing.T) {
	repo := memory.NewUserRepo()
	repo.Save(&domain.User{ID: "u1", Username: "admin", Role: domain.RoleAdmin})

	u, ok := repo.GetByUsername("admin")
	if !ok {
		t.Fatal("expected to find admin user")
	}
	if u.Role != domain.RoleAdmin {
		t.Errorf("expected role admin, got %q", u.Role)
	}
}

func TestUserRepo_NotFound(t *testing.T) {
	repo := memory.NewUserRepo()
	_, ok := repo.GetByUsername("ghost")
	if ok {
		t.Error("expected not found")
	}
}

func TestUserRepo_OverwriteOnSave(t *testing.T) {
	repo := memory.NewUserRepo()
	repo.Save(&domain.User{ID: "u1", Username: "admin", Role: domain.RoleAdmin})
	repo.Save(&domain.User{ID: "u1", Username: "admin", Role: domain.RoleUploader})

	u, _ := repo.GetByUsername("admin")
	if u.Role != domain.RoleUploader {
		t.Error("expected second save to overwrite role")
	}
}
