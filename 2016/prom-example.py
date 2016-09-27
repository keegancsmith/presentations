from prometheus_client import Counter
c = Counter('my_failures_total', 'Description of counter')
c.inc()     # Increment by 1
c.inc(1.6)  # Increment by given value

# There are utilities to count exceptions raised:
@c.count_exceptions()
def f():
    pass

with c.count_exceptions():
    pass

# Can have labels
c = Counter('github_requests_total', 'Total requests to GH API', ['type', 'code'])
c.labels('repos', '403').inc()
c.labels('user', '200').inc()
