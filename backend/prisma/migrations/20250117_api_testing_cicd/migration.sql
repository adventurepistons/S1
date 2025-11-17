-- CreateEnum
CREATE TYPE "ApiTestType" AS ENUM ('REST', 'GRAPHQL');

-- CreateEnum
CREATE TYPE "WebhookProvider" AS ENUM ('GITHUB', 'GITLAB', 'BITBUCKET');

-- CreateEnum
CREATE TYPE "NotificationType" AS ENUM ('SLACK', 'TEAMS', 'DISCORD', 'WEBHOOK', 'EMAIL');

-- CreateTable
CREATE TABLE "api_test_definitions" (
    "id" TEXT NOT NULL,
    "project_id" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "description" TEXT,
    "type" "ApiTestType" NOT NULL,
    "method" TEXT,
    "url" TEXT NOT NULL,
    "headers" JSONB,
    "body" JSONB,
    "query" TEXT,
    "variables" JSONB,
    "assertions" JSONB,
    "timeout" INTEGER NOT NULL DEFAULT 30000,
    "is_active" BOOLEAN NOT NULL DEFAULT true,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "api_test_definitions_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "api_test_results" (
    "id" TEXT NOT NULL,
    "test_id" TEXT NOT NULL,
    "execution_id" TEXT,
    "passed" BOOLEAN NOT NULL,
    "duration" INTEGER NOT NULL,
    "status_code" INTEGER,
    "response_body" JSONB,
    "response_headers" JSONB,
    "error_message" TEXT,
    "assertions" JSONB,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "api_test_results_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "webhook_configs" (
    "id" TEXT NOT NULL,
    "project_id" TEXT,
    "user_id" TEXT NOT NULL,
    "provider" "WebhookProvider" NOT NULL,
    "webhook_url" TEXT NOT NULL,
    "secret" TEXT NOT NULL,
    "events" JSONB NOT NULL,
    "is_active" BOOLEAN NOT NULL DEFAULT true,
    "last_triggered" TIMESTAMP(3),
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "webhook_configs_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "notification_configs" (
    "id" TEXT NOT NULL,
    "project_id" TEXT,
    "user_id" TEXT NOT NULL,
    "type" "NotificationType" NOT NULL,
    "webhook_url" TEXT NOT NULL,
    "events" JSONB NOT NULL,
    "is_active" BOOLEAN NOT NULL DEFAULT true,
    "metadata" JSONB,
    "last_sent" TIMESTAMP(3),
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "notification_configs_pkey" PRIMARY KEY ("id")
);

-- CreateIndex
CREATE INDEX "api_test_definitions_project_id_idx" ON "api_test_definitions"("project_id");

-- CreateIndex
CREATE INDEX "api_test_definitions_type_idx" ON "api_test_definitions"("type");

-- CreateIndex
CREATE INDEX "api_test_results_test_id_idx" ON "api_test_results"("test_id");

-- CreateIndex
CREATE INDEX "api_test_results_execution_id_idx" ON "api_test_results"("execution_id");

-- CreateIndex
CREATE INDEX "api_test_results_created_at_idx" ON "api_test_results"("created_at" DESC);

-- CreateIndex
CREATE INDEX "webhook_configs_user_id_idx" ON "webhook_configs"("user_id");

-- CreateIndex
CREATE INDEX "webhook_configs_provider_idx" ON "webhook_configs"("provider");

-- CreateIndex
CREATE INDEX "notification_configs_user_id_idx" ON "notification_configs"("user_id");

-- CreateIndex
CREATE INDEX "notification_configs_type_idx" ON "notification_configs"("type");

-- AddForeignKey
ALTER TABLE "api_test_definitions" ADD CONSTRAINT "api_test_definitions_project_id_fkey" FOREIGN KEY ("project_id") REFERENCES "projects"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "api_test_results" ADD CONSTRAINT "api_test_results_test_id_fkey" FOREIGN KEY ("test_id") REFERENCES "api_test_definitions"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "webhook_configs" ADD CONSTRAINT "webhook_configs_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "notification_configs" ADD CONSTRAINT "notification_configs_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users"("id") ON DELETE CASCADE ON UPDATE CASCADE;
