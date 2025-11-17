import {
  WebSocketGateway,
  WebSocketServer,
  SubscribeMessage,
  OnGatewayConnection,
  OnGatewayDisconnect,
  MessageBody,
  ConnectedSocket,
} from '@nestjs/websockets';
import { Server, Socket } from 'socket.io';
import { Logger } from '@nestjs/common';
import { CacheService } from '../shared/cache.service';

@WebSocketGateway({
  namespace: '/tests',
  cors: {
    origin: '*',
    credentials: true,
  },
})
export class TestExecutionGateway implements OnGatewayConnection, OnGatewayDisconnect {
  @WebSocketServer()
  server: Server;

  private readonly logger = new Logger(TestExecutionGateway.name);
  private subscriptions: Map<string, any> = new Map();

  constructor(private cacheService: CacheService) {}

  handleConnection(client: Socket) {
    this.logger.log(`Client connected: ${client.id}`);
  }

  handleDisconnect(client: Socket) {
    this.logger.log(`Client disconnected: ${client.id}`);

    // Clean up subscriptions
    this.subscriptions.delete(client.id);
  }

  /**
   * Subscribe to test execution logs
   * Client sends: { executionId: 'exec_123' }
   */
  @SubscribeMessage('stream')
  handleStream(
    @MessageBody() data: { executionId: string },
    @ConnectedSocket() client: Socket,
  ) {
    const { executionId } = data;

    this.logger.log(`Client ${client.id} subscribing to execution: ${executionId}`);

    // Join room for this execution
    client.join(`execution:${executionId}`);

    // Subscribe to Redis pub/sub for logs
    this.cacheService.subscribe(`logs:${executionId}`, (message) => {
      // Emit to all clients in this execution's room
      this.server.to(`execution:${executionId}`).emit('log', message);
    });

    // Store subscription for cleanup
    this.subscriptions.set(client.id, executionId);

    // Send confirmation
    client.emit('subscribed', {
      executionId,
      message: `Subscribed to execution ${executionId}`,
    });
  }

  /**
   * Unsubscribe from execution
   */
  @SubscribeMessage('unsubscribe')
  handleUnsubscribe(
    @MessageBody() data: { executionId: string },
    @ConnectedSocket() client: Socket,
  ) {
    const { executionId } = data;

    this.logger.log(`Client ${client.id} unsubscribing from execution: ${executionId}`);

    // Leave room
    client.leave(`execution:${executionId}`);

    // Remove subscription
    this.subscriptions.delete(client.id);

    client.emit('unsubscribed', { executionId });
  }
}
