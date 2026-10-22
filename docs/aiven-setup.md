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
