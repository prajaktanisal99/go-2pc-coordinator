Q: Redis doesn't support 2PC, so how did you include it in your transaction?
A: I implemented a Lock-Based Adapter. During the Prepare phase, I used Redis SETNX to acquire a distributed lock on the specific resource. This effectively 'reserved' the data. In the Commit phase, I used a Redis Pipeline to atomically apply the data change and release the lock. If the Prepare phase failed for Postgres, the Coordinator would call Rollback on Redis, which simply released the lock.

-----
