## Filtering Overhead
```
optimized-db=# EXPLAIN Analyze SELECT * FROM fire_calls where City = 'SF';
                                                    QUERY PLAN                                                 
--------------------------------------------------------------------------------------------------------------------
 Seq Scan on fire_calls  (cost=0.00..8718.20 rows=120317 width=261) (actual time=0.098..53.212 rows=120072 loops=1)
   Filter: ((city)::text = 'SF'::text)
   Rows Removed by Filter: 55224
 Planning Time: 0.656 ms
 Execution Time: 60.163 ms
(5 rows)

optimized-db=# EXPLAIN Analyze SELECT * FROM fire_calls;
                                                     QUERY PLAN                                                   
--------------------------------------------------------------------------------------------------------------------
 Seq Scan on fire_calls  (cost=0.00..8279.96 rows=175296 width=261) (actual time=0.022..19.618 rows=175296 loops=1)
 Planning Time: 0.143 ms
 Execution Time: 29.343 ms
(3 rows)
```
- Look at the first one (filter), and the second one (no filter), the one with filter is slower.
- Filtering has overhead.

## 