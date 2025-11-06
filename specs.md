# UAE cost of living

the idea is simple, build a simple website that calculates the cost of living to users. This is the frontend. 

Before we proceed let's plan first:
- planning should be not sprint like, we use agents and sometimes parallel ones
- avoid slop.
- complexity is always terrible
- being pragmatic is a must
- bad code that is readable, and can work is okay
- verbose code is not okay
- optimize for debugability and readability


## backend
- we will utilize durable workflow (temporal) that will pull in data resources from various places such as bayut, dubizzle, etc (the merier the better) -> we need to design a data model to know what we are grabbing and what we are saving
- we will have a workflow that will pull in from these resources and compensate as we go (this will run indefinitely as we alwyas need to update our db and compensate whenever listing had changed). i know temporal is an overkill but we can make it less painful (we can use sqlite if that is possible)
- deployment: we must establish a clear docker-compose that works
- i need grafana and otel : this is a must
- backend is go
- frontend is templ + htmx + html
