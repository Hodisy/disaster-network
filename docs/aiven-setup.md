# Aiven Runtime setup

Repository: https://github.com/Hodisy/disaster-network

## Scope

The root `Dockerfile` prepares the existing mock-only Nuxt frontend for Aiven Runtime. It installs dependencies with Bun and the committed lockfile, builds Nuxt, and runs the generated Nitro application with Node 24 as a non-root user on port 8080.

No business backend, database adapter, real voice session, or external API integration is added. The existing PostgreSQL service is not required by the current frontend. Do not provision another database for this setup.

The container has not been built or run as part of this change. Application deployment and connection tests remain pending.

## Console steps

1. Commit and push the reviewed application and deployment files to the intended branch of `Hodisy/disaster-network`. Local changes are not visible to Aiven until pushed.
2. Open the Aiven project, choose **Runtime**, then **Deploy application**.
3. Connect GitHub if needed. Review the installation permissions and select only the intended repository.
4. Select `Hodisy/disaster-network` and the intended branch.
5. Select the root `Dockerfile` as the manifest and scan it. Use the repository root as the build context and HTTP port **8080**.
6. Review the region, resource plan, credit usage, and estimated cost before deployment. No PostgreSQL connection or AI API key is needed for the mock frontend.
7. Deploy when authorized. Open `/hotline`, `/client`, and `/insurer` on the generated HTTPS address, and inspect the application logs if startup fails.

Source: [Aiven deployment documentation](https://aiven.io/docs/products/runtime/deploy-apps) and [manifest documentation](https://aiven.io/docs/products/runtime/manifest-files/manifests).

## Later database integration

When a separate backend phase is authorized, reuse the existing PostgreSQL service `pg-3351099e` in project `marin-f90a`. In the deployed application's **Connected services** section, select **Connect service** and the existing database. Confirm the connection variable used by the future server adapter, such as `DATABASE_URL`.

Database credentials belong exclusively on the server. Never place them in public Nuxt configuration or frontend code. The application does not currently read `DATABASE_URL`; adding a connection variable alone does not implement persistence. Migrations, TLS verification, connection pooling, and durable document storage require that later backend work.

Source: [Connect services to Aiven Runtime](https://aiven.io/docs/products/runtime/connect-services-to-apps).

## Local PostgreSQL load - 22 October 2026

The user subsequently authorized connecting to the existing service and loading the small Jules dataset. The ignored local `.env` now contains the connection URL, CA path, reset feature flag and maintenance key. Verified TLS connectivity and `db:setup` succeeded. Read-only counts: nine tables, three customers/policies/claim records, four documents and eleven sections, all inside `disaster_demo`. No other schema was changed.

The user subsequently authorized launching Nuxt locally and testing Settings. The reset succeeded against Aiven and the browser confirmed three restored customers, policies and claim records. Shared database modules are explicitly bundled by Nitro so compiled imports resolve correctly. No production build was run. The hosted application's connected services/environment have not been changed by this load task; `.env` is local and excluded from Git and Docker context. A hosted server requires the corresponding private runtime variables and an available CA certificate independently of this local configuration.

## Application update — 22 October 2026

The subsequent release request authorizes pushing the current hotline/evidence/insurer implementation and generic model instructions to `main`, and updating the existing `application-8agak1` Runtime. The production build passed locally. The Docker context now includes the incident-context JSON imported by Nitro; unrelated design/submission documents remain outside it. Do not reset or seed the database during deployment.

Runtime needs the existing PostgreSQL `DATABASE_URL` plus the project CA in private `DATABASE_CA_CERT`, the existing private OpenAI key and hotline access/origin settings. Local `.env` files are never committed or automatically deployed. Insurer analysis additionally needs its separate private gate and enabled flag described in `insurer-agent.md`. Use the exact public browser origin (`https://getdisasternetwork.com`) for hotline and insurer origin checks. Connecting a database or changing runtime variables is distinct from building/pushing code; confirm the actual configuration before claiming remote AI availability.
