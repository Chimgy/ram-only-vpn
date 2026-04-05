import { Injectable, ConflictException } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { randomBytes } from 'crypto';
import { User } from '../users/user.entity';

@Injectable()
export class AuthService {
  constructor(
    @InjectRepository(User)
    private readonly users: Repository<User>,
  ) {}

  async register(): Promise<{ user_id: string; valid_until: Date }> {
    // Retry on the rare collision
    for (let attempt = 0; attempt < 5; attempt++) {
      const user_id = this.generateUserId();
      const existing = await this.users.findOneBy({ user_id });
      if (existing) continue;

      const valid_until = new Date();
      valid_until.setDate(valid_until.getDate() + 30);

      const user = this.users.create({ user_id, valid_until });
      await this.users.save(user);

      return { user_id, valid_until };
    }

    throw new ConflictException('Failed to generate unique ID, try again');
  }

  // Cryptographically random 16-digit number (1000000000000000 – 9999999999999999)
  private generateUserId(): string {
    const bytes = randomBytes(8);
    const big = bytes.readBigUInt64BE();
    const id = (big % 9000000000000000n) + 1000000000000000n;
    return id.toString();
  }
}
