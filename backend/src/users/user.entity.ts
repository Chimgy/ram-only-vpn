import { Entity, PrimaryColumn, Column } from 'typeorm';

@Entity('users')
export class User {
  @PrimaryColumn({ type: 'varchar', length: 16 })
  user_id: string;

  @Column({ type: 'timestamptz' })
  valid_until: Date;
}
