package queue

import (
	"log"
	"main/internal/server/service"
	"main/pkg/solana"
	"time"
)

func AddWalletToQueue(walletAddress string) chan Result {
	queueMapMutex.Lock()
	defer queueMapMutex.Unlock()

	newChan := make(chan Result, 1)

	if _, exists := queueMap[walletAddress]; !exists {
		queueStack := []*chan Result{&newChan}
		queueMap[walletAddress] = queueStack
		go runWalletQueue(walletAddress)
	} else {
		queueMap[walletAddress] = append(queueMap[walletAddress], &newChan)
	}
	// cleaning up the channel and queue to avoid memory leaks
	go func() {
		<-time.After(30 * time.Second)
		queueMapMutex.Lock()
		defer queueMapMutex.Unlock()
		if chans, exists := queueMap[walletAddress]; exists {
			for i, ch := range chans {
				if ch == &newChan {
					close(*ch)
					queueMap[walletAddress] = append(chans[:i], chans[i+1:]...)
					break
				}
			}

			if len(queueMap[walletAddress]) == 0 {
				delete(queueMap, walletAddress)
			}
		}
	}()

	return newChan
}

func popJobFromWalletQueue(walletAddress string) *chan Result {
	queueMapMutex.Lock()
	defer queueMapMutex.Unlock()

	if chans, exists := queueMap[walletAddress]; exists && len(chans) > 0 {
		poppedChan := chans[0]
		queueMap[walletAddress] = chans[1:]
		if len(queueMap[walletAddress]) == 0 {
			delete(queueMap, walletAddress)
			return poppedChan
		}
	}
	return nil
}

func runWalletQueue(walletAddress string) {
	for {
		queueMapMutex.RLock()
		val, exists := queueMap[walletAddress]
		queueMapMutex.RUnlock()
		if exists && len(val) > 0 {
			if amount, err := service.GetWallet(walletAddress); err == nil {
				*val[0] <- Result{Result: amount, Error: nil, Cache: true}
				popJobFromWalletQueue(walletAddress)
				continue
			}

			amount, err := solana.NewSolClient().GetBalance(walletAddress)
			if err != nil {
				// sending errors to all channels in the queue
				for {
					ch := popJobFromWalletQueue(walletAddress)
					if ch == nil {
						break
					}
					*ch <- Result{Result: "", Error: err, Cache: false}
				}
			}
			*val[0] <- Result{Result: amount, Error: nil, Cache: false}
			err = service.SetWallet(walletAddress, amount)
			if err != nil {
				log.Println("Error setting wallet to cache:", err)
			}
			popJobFromWalletQueue(walletAddress)
			continue
		} else {
			break
		}
	}
}
