import { IsString, IsArray, IsOptional, IsBoolean, IsInt, Min, Max, IsEnum } from 'class-validator';

export enum BrowserType {
  CHROME = 'chrome',
  FIREFOX = 'firefox',
  SAFARI = 'safari',
  EDGE = 'edge',
}

export enum EnvironmentType {
  DEV = 'dev',
  STAGING = 'staging',
  PRODUCTION = 'production',
}

export class ExecuteTestsDto {
  @IsString()
  projectId: string;

  @IsArray()
  @IsString({ each: true })
  @IsOptional()
  tests?: string[]; // Test names or "*" for all

  @IsEnum(EnvironmentType)
  @IsOptional()
  environment?: EnvironmentType = EnvironmentType.DEV;

  @IsEnum(BrowserType)
  @IsOptional()
  browser?: BrowserType = BrowserType.CHROME;

  @IsBoolean()
  @IsOptional()
  headless?: boolean = true;

  @IsBoolean()
  @IsOptional()
  parallel?: boolean = false;

  @IsInt()
  @Min(1)
  @Max(10)
  @IsOptional()
  maxWorkers?: number = 1;

  @IsInt()
  @Min(0)
  @Max(5)
  @IsOptional()
  retryFailedTests?: number = 0;

  @IsInt()
  @Min(5000)
  @Max(300000)
  @IsOptional()
  timeout?: number = 30000; // 30 seconds

  @IsArray()
  @IsString({ each: true })
  @IsOptional()
  tags?: string[]; // e.g., ['smoke', 'critical']
}
