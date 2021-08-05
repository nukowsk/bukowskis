Auction v1
==========

## Changelog
27.01.2021 - initial draft

## Context

Bukowskis is a service that conducts auctions for streams of
transactions submitted by wallets. Bidders can submit bids for one or
series of target block heights. The highest bid wins. When the target
block arrives, all transactions received until the next block arrives
will be relayed to an address provided by the winning bidder. 

### Payments
In order to process payments, Bukowskis will control accounts on
Ethereum for each bidder. Bidders can top up the account at any
time. At the end of the block in which a bidder has won, the bid amount
will be deducted from the appropriate account and pooled into a common
account for later disbursement to users in the form of a [merkle
distribution](https://github.com/Uniswap/merkle-distributor)

## Considerations
In order to ensure reliable delivery of transactions which bidders have
paid for, Bukowskis will run in the cloud with redundant instances
receiving a partition of traffic.  The state of
the auctions must be stored durably and provide a consistent state of the
auction across instances of Bukowskis.

Additionally, payments are initiated from Bukowskis but can only be
settled when transactions are confirmed by the blockchain. Therefore
bukowskis must keep track of pending payments and accurate account
balances even when instances of Bukowskis crash or is restarted during
deployment.

## Design

### Data structures

Data will be stored in Google Firestore. Transactions will be used to
ensure atomic operations and consistent view of state across instances
of Bukowskis.

### Auctions
Auctions are conducted per source of the transaction stream.

```
type Auction struct {
    height: ...
    state: [open, closed, settled]
    confirmation: txHash
}

// storage, indexed by height
${source}/
    auctions/
        #{height}: Auction{...}
```

### Bids
```
type Bid struct {
    height: ...
    bidder_id: ...
    amount: ...
}

#{source}/
    bids/
        #{height}/
            #{bid_id}: Bid{ ... }
```

### Payments

```
type Payment struct  {
    id: ...
    bid: Bid { }
    status: [pending, confirmed, disbursed]
    confirmation: txHash
}

#{source}/
    payments/
        #{height}/
            #{payment_id}: Payment {...},
            ...,
```

### State Machine
The core logic of Bukowskis will be modeled as a state machine which
will respond to events. Events will come in from multiple sources
including subscription to full nodes and the Bukowskis http endpoint but
will be funneled into a single state machine that will facilitate
deterministic testing. 

Source - Events:
* Endpoint - Transaction
* Endpoint - Bid
* Blockchain - NewBlock
* Blockchain - Settlment

### Event Handling
```
type Auction struct {
    ...
}

func (a ...) hanldeBid(bid) {
    // update bid and set winner for the block
}

func (a ...) handleTransaction(transaction) url {
    // return the url 
}

func (a ...) handleNewBlock(block) paymentTx {
    // initiate payment
}

func (a ...) handleNewPayment(...) {
    // settlement payment
}

func (a ...) process(event) result {
    // demux the event to the handler and produce a result
}
```

### Testing
```
// auctions are per source stream
auction := NewAuction(...)

events := []event {
    NewBlock{...},
    Bid{bidder: 1, ...},
    Transaction{...},
    Bid{bidder: 2, ...},
    Bid{bidder: 3...},
    Transaction{...},
    NewBlock{...},
}

// Apply a sequence of events to the auction
result := auction.process(events)

//  assert the balances after the auction has been conducted
assert(result.Balances, map[d]balance{
    1: ...,
    2: ...,
})
```

### Payment Processing
Payments stored in firestore reflect the most up date account balances
for bidders based on auction currently executing. Settlement
transactions will be submmited to the blockchain and include an
identifier in the memo to link the transactions with the record in
firestore. In cases in which the process crashes or it takes several
attempts for the transaction to get accepted by the blockchain with
multiple crashes in between.

## Consequences
* Bukowskis provides a reliable auction service that settles payments
  fairly even under process instability
* Bukowskis will synchronize state between processes, firestore and
  the ethereum blockchain in a way minimizes interference with running
  blockchains
* Payment will be processed consistently even when it takes a few blocks
  for them to be confirmed

