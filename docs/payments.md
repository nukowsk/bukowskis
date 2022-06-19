Payments
========
We need to collect fees from bidders for auctions they win.
Those fees should be pooled into an account that can be sourced for
merkle drops by a manual process.

We need to ensure that bidders pay reliably, even when the bukowskis,
firebase or ethereum operations fail.

## Constraints
1. Data storage is ethereum is expensive
   We can't store all the neccessary data in ethereum due to both cost of
   storage and transactions costs of writting to ethereum

   1.1 We don't want to initiate ethereum transactions for every auction

2. Processes running concurrently require a consistent view of bidder
   account balances.

   2.1 Conflicting bids sent to multiple processes should produce
   consistent highest bid wins outcomes
    
3. Operations between firebase and ethereum are not transactional

   3.1 At any point in the process, Writes to ethereum can fail when writes to firebase suceeed and
   vise versa


### Data Structure

```

type Payment struct  {
    height: ...
    id: bidderID
    amount: ...
    status: [pending, confirmed, disbursed]
}

#{height}/
    #{payment_id}: Payment {...},
            ...,


poolAccount:
    #{txUD}:
        from: #{bidderAccount},
        ...
        data: #{settledHeight},
    ...
        
```
## Account Balances
A bidders account balance is equal to their eth account balance - any
pending payments.

Balances are from firebase at initiation and bids are initiated with
compare and swap operations against account balances.

```
bidders/
    #{bidderID}/
        {
            maxSettled: ..., // last settled height?
            balance: amount,
            pending:  amount,
        }
```
## Making a Bid

TODO: Do this as a transaction that ensure that the bidder has
sufficient balance

## Payment Initiation

Bidders should pay for blocks in which they have the highest bid after
the first transaction eligible for that block is forwarded to them.

pre-condition: The bid is valid, the bidder had sufficient balance to
submit the bid
```
actionRef := client.Collection("auctions").Doc(height)
err := client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
        auction, err := tx.Get(ref)
        if err != nil {
                return err
        }
        bids, err := auction.Collection("bids")
        if err != nil {
                return err
        }

        max := 0
        winner := nil
        for bid := range bids {
            if bid.amount > max {
                winner = bid
                max = bid.amount
            }
        }

        // Now we need to subtract the balance
        bidderRef := client.Collection("bidders").Doc(bid.bidderID)
        if err != nil {
            return err
        }

        bidder, err := tx.Get(bidderRef)
        if err != ni {
            return err
        }

        // XXX: assumption, amount reflects the balance on ethereum
        pending := bidder.pending
        tx.Set(bidderRef, map[string]interface{}{
                "pending": pending.(int64) + bid.amount,
        }, firestore.MergeAll)

        // TODO: Create a payment transaction
})
```

### Settlement Process
A settlement process can be run  manually to settle pending payments by
synchronizing the state of firebase with ethereum.

1. get the `map[bidderID]maxSettled` as `maxSettledFirebase` from firebase
   let `min(maxSettledFirebase)` as `minMaxSettledFirebase`
   
2. get max settled height from ethereum
   Collect all transactions on poolAccount where height > `minMaxSettledFirebase`
   let `map[bidderID]maxSettled` as `maxSettledEthereum`

NOTE: New bidders might not have settled on ethereum yet and should
therefore have maxSettled=0

3. Update firebase payments marked as pending but have already been
   settled on ethereum

This happends when the settlement processes crashes after submitting
transactions to ethereum but before recording the update in firebase

```
for bidderID, maxSettled := range maxSettledEthereum {
    bidderRef, err := client.Collection("bidders").Doc(bidderID)
    if err != nil {
        // edge case?
    }
    err := client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
        bidder := tx.Get(bidderRef)

        /*
        What if the settlement process crashes, and then ethMaxSettled
        is at h. but then the this bidder wins h+1, so in effect we have
        pending > 0 for transactions after ethMaxSettled
        so this scheme is missing something
        */
        if bidder.MaxSettled < maxSettled {
            // Confirm that we are settled up to maxSettled
            return tx.Set(bidderRef, map[string]interface{}{
                    "pending": 0,
                    "maxSettled": maxSettled
            })
        }
    }
}
```
    
2. Get the pending payments from firebase as `firebasePending`
```
SELECT bidderID, max(height) as maxHeight, sum(amount) as amount
FROM payments 
WHERE
    state == 'pending'
GROUP BY
    bidderID
```


4. Sync ethereum with firebase
```
for bidderID, maxHeight, amount in firebasePending {
    ethereum.send(
        to: bidderID,
        amount: amount,
        data: maxHeight,
    )
}
```

Pre-condition: step #4 succeeded, duh
5. Update firebase
```
for bidderID, maxHeight, _ in firebasePending {
    firebase.update(
        UPDATE payments
        WHERE 
            bidderID = BidderID
        AND state = pending
        AND height <= maxHeight
        SET
            state=settled
    )
}
```
