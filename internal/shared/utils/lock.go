package utils

import (
	"sync"

	"github.com/google/uuid"
)

type WalletLockManager struct {
	locks sync.Map
}

func NewWalletLockManager() *WalletLockManager {
	return &WalletLockManager{}
}

func (walletLockManager *WalletLockManager) GetLock(walletID uuid.UUID) *sync.Mutex {
	lock, _ := walletLockManager.locks.LoadOrStore(walletID.String(), &sync.Mutex{})
	return lock.(*sync.Mutex)
}

func (walletLockManager *WalletLockManager) LockWallet(walletID uuid.UUID) {
	lock := walletLockManager.GetLock(walletID)
	lock.Lock()
}

func (walletLockManager *WalletLockManager) UnlockWallet(walletID uuid.UUID) {
	lock := walletLockManager.GetLock(walletID)
	lock.Unlock()
}

func (walletLockManager *WalletLockManager) LockWallets(walletIDs ...uuid.UUID) {
	// Sort wallet IDs to prevent deadlocks
	sortedIDs := make([]uuid.UUID, len(walletIDs))
	copy(sortedIDs, walletIDs)

	// Simple sort by string representation
	for i := 0; i < len(sortedIDs); i++ {
		for j := i + 1; j < len(sortedIDs); j++ {
			if sortedIDs[i].String() > sortedIDs[j].String() {
				sortedIDs[i], sortedIDs[j] = sortedIDs[j], sortedIDs[i]
			}
		}
	}

	// Lock in sorted order
	for _, id := range sortedIDs {
		walletLockManager.LockWallet(id)
	}
}

func (walletLockManager *WalletLockManager) UnlockWallets(walletIDs ...uuid.UUID) {
	// Unlock in reverse order
	for i := len(walletIDs) - 1; i >= 0; i-- {
		walletLockManager.UnlockWallet(walletIDs[i])
	}
}
