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

## Index vs No Index Performance and EXPLAIN Output
```
optimized-db=# EXPLAIN ANALYZE SELECT * FROM fire_calls WHERE call_type = 'Medical Incident';
                                                     QUERY PLAN                                                    
--------------------------------------------------------------------------------------------------------------------
 Seq Scan on fire_calls  (cost=0.00..8718.20 rows=114065 width=261) (actual time=0.047..65.169 rows=113794 loops=1)
   Filter: ((call_type)::text = 'Medical Incident'::text)
   Rows Removed by Filter: 61502
 Planning Time: 0.298 ms
 Execution Time: 71.071 ms
(5 rows)

CREATE INDEX idx_fire_calls_calltype ON fire_calls (call_type);

optimized-db=#  EXPLAIN ANALYZE SELECT * FROM fire_calls WHERE call_type = 'Medical Incident';
                                                     QUERY PLAN                                                   
--------------------------------------------------------------------------------------------------------------------
 Seq Scan on fire_calls  (cost=0.00..8718.20 rows=114065 width=261) (actual time=0.041..74.508 rows=113794 loops=1)
   Filter: ((call_type)::text = 'Medical Incident'::text)
   Rows Removed by Filter: 61502
 Planning Time: 0.306 ms
 Execution Time: 80.080 ms
(5 rows)

```
- The index was not even used, this is because Medical Incident occurred too many times, hence making it inefficient. It was faster to count 
sequentially.

```
optimized-db=# EXPLAIN ANALYZE SELECT * FROM fire_calls WHERE call_type = 'Vehicle Fire';
                                                         QUERY PLAN                                                          
-----------------------------------------------------------------------------------------------------------------------------
 Gather  (cost=1000.00..8524.10 rows=841 width=261) (actual time=0.725..44.282 rows=854 loops=1)
   Workers Planned: 2
   Workers Launched: 2
   ->  Parallel Seq Scan on fire_calls  (cost=0.00..7440.00 rows=350 width=261) (actual time=0.563..14.100 rows=285 loops=3)
         Filter: ((call_type)::text = 'Vehicle Fire'::text)
         Rows Removed by Filter: 58147
 Planning Time: 0.276 ms
 Execution Time: 44.393 ms
(8 rows)

CREATE INDEX idx_fire_calls_calltype ON fire_calls (call_type);

optimized-db=#  EXPLAIN ANALYZE SELECT * FROM fire_calls WHERE call_type = 'Vehicle Fire';
                                                             QUERY PLAN                                                              
-------------------------------------------------------------------------------------------------------------------------------------
 Bitmap Heap Scan on fire_calls  (cost=10.94..2359.36 rows=841 width=261) (actual time=0.628..3.520 rows=854 loops=1)
   Recheck Cond: ((call_type)::text = 'Vehicle Fire'::text)
   Heap Blocks: exact=777
   ->  Bitmap Index Scan on idx_fire_calls_calltype  (cost=0.00..10.73 rows=841 width=0) (actual time=0.434..0.436 rows=854 loops=1)
         Index Cond: ((call_type)::text = 'Vehicle Fire'::text)
 Planning Time: 0.659 ms
 Execution Time: 3.990 ms
(7 rows)
```
- In this case, the index made the query significantly faster. 
- You can see that the index was called in this example by looking at the `Bitmap Index Scan`
