# Coding Challenge: Balance the Loan Books

## Execution
Ensure that CSV files are in "large" folder then run:

```
go run main.go
```

Output will be placed in the same directory as "assignments.csv" and "yields.csv" files.


## Answers:

1. Assignment took 4 hours. Loan matching was done in a brute force manner, with no consideration for optimizing yields.
2. In expanded model new covenants could be introduced as an encoded list of conditions. For example: ("!=", "CityName"), (">", "MinYield"). The list of operations could be dynamically evaluated at run-time. For even more complex rules, covenants could be stored as functions.
3. A separate data store could be used to maintain a list of facilities. A service for adding new facilities, or returning an ordered list of suitable facilities could be wrapped around the data store.
4. REST API endpoints defined under root (i.e. /api/v1/):
    - /loans/{loan_id} -- POST assign loan to facility id (provided in JSON payload), GET for status query
    - /facilities/{facility_id} - GET provides capacity
5. For batch assignments the loans could be ordered with the largest first. Similarly to memory fragmentation problem, smaller chunks are easier to allocate than larger ones.
6. Run-time complexity of the loan assignment portion is O(facilities * covenants), a.k.a "atrocious." To optimize, both facilities and covenants have to be ordered by fields which are most important for the assignment algorithm.
