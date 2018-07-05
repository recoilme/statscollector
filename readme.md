# Statcollector

```
// View register urls view
// Example:
// curl -d '{"referer":"hotpop","urls":["url1","url2"]}' -H "Content-Type: application/json" -X POST http://localhost:8088/api/view


// Click register urls clicks
// Example:
// curl -d '{"referer":"hotpop","urls":["url1","url2"]}' -H "Content-Type: application/json" -X POST http://localhost:8088/api/click


// Stat show stat
// Example:
// curl  -H "Content-Type: application/json" -X GET http://localhost:8088/api/stat/hotpop
// [{"Url":"url1","View":3,"Click":2,"CTR":1.5},{"Url":"url2","View":3,"Click":2,"CTR":1.5}]


```