# docker-auto-cleaner

A tool to automatically cleanup residual docker artifacts

# Testing
**!!! Attention !!!**
Be careful when running the tests, as they will remove all docker artifacts that are not in use.

# Requirements

- Kill containers running longer than a specified duration
- Remove containers that have exited
- Remove images haven't been used in a specified duration
- Remove volumes that are not attached to any containers
- Remove networks that are not attached to any containers
- Remove the least recently used images if the total size of images exceeds a specified value
