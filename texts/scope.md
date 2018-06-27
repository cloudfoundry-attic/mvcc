## Objective of this test suite:

### This test suite shall:
- Cover positive cases as follows:
  - Users with new fine-grained permissions (FGPs) from PERM (for example: `space.create`)
  - Test other pseudo-authorization cases in CC (for example: suspended orgs)
- Cover negative cases as follows:
  - Users with FGPs in other orgs/spaces
  - Authorized users with no scopes, roles, or FGPs
- Integrate a running CC with a running PERM
- Test using HTTP requests to CC v2 and v3 APIs and the PERM API
- Mock out all other external collaborators (for example: UAA)

### This test suite shall NOT:
- Backfill integration testing for roles in the CCDB
- Backfill integration testing for scopes (`admin`/ `read_only_admin`/ `global_auditor`)
- Cover negative cases as follows:
  - Unauthorized users: no/invalid UAA token
  - Authorized users without `cloud_controller.read` or `cloud_controller.write` scopes
  - Users with other roles in CCDB (for example: creating space as OrgAuditor/OrgMember)
  - Users with read only UAA scopes (`read_only_admin`/ `global_auditor`)

--- 

### Why this test suite exists
- Faster than CATs because it doesn't require a bosh deploy
- Tests CC & Perm "for realz" because we don't cheat by bypassing APIs to seed databases
- Test run faster than RSpec integration tests in CC ????? -- underlying question: why not the perm integration tests in CC?


### Fears
- Mocking out all of the external CC integrations seems hard (running apps on a fake Diego?) 


### Other related suites
- CATs: Test behavior of CC endpoints, but much less exhaustive permissions testing
- Perm integration rspec suite in cloud controller: Test CC role writes are propagated to perm
