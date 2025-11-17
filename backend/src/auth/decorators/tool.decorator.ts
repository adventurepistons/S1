import { SetMetadata } from '@nestjs/common';

export const Tool = (toolName: string) => SetMetadata('tool', toolName);
