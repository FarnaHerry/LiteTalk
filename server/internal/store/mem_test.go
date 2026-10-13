package store

import (
	"context"
	"testing"
)

// TestMemStore 内存实现基本 CRUD（语义与 MySQL 对齐的回归保障）。
func TestMemStore(t *testing.T) {
	ctx := context.Background()
	m := NewMem()

	// users 唯一性。
	u1 := &User{Email: "a@ex.com", Nickname: "alice", Tag: "0123", PasswordHash: "h"}
	if err := m.CreateUser(ctx, u1); err != nil {
		t.Fatal(err)
	}
	if err := m.CreateUser(ctx, &User{Email: "a@ex.com", Nickname: "x", Tag: "9999"}); err != ErrConflict {
		t.Fatalf("邮箱冲突应报 ErrConflict，实际 %v", err)
	}
	if err := m.CreateUser(ctx, &User{Email: "b@ex.com", Nickname: "alice", Tag: "0123"}); err != ErrConflict {
		t.Fatalf("(nickname,tag) 冲突应报 ErrConflict，实际 %v", err)
	}

	// friendships 双向唯一。
	f := &Friendship{UserID: u1.ID, FriendID: 2, Status: FriendPending}
	if err := m.CreateFriendship(ctx, f); err != nil {
		t.Fatal(err)
	}
	if err := m.CreateFriendship(ctx, &Friendship{UserID: 2, FriendID: u1.ID}); err != ErrConflict {
		t.Fatalf("反向关系应冲突，实际 %v", err)
	}

	// servers + members + channels + messages 级联删除。
	s := &Server{Name: "s", OwnerID: u1.ID, IconName: "music", IconColor: "#4FA4F3"}
	if err := m.CreateServer(ctx, s); err != nil {
		t.Fatal(err)
	}
	if err := m.AddMember(ctx, &ServerMember{ServerID: s.ID, UserID: u1.ID, Role: RoleOwner}); err != nil {
		t.Fatal(err)
	}
	ch := &Channel{ServerID: s.ID, Name: "c", Type: ChannelText}
	if err := m.CreateChannel(ctx, ch); err != nil {
		t.Fatal(err)
	}
	if err := m.CreateMessage(ctx, &Message{ChannelID: ch.ID, UserID: u1.ID, Content: "hi"}); err != nil {
		t.Fatal(err)
	}
	if err := m.DeleteServer(ctx, s.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := m.ChannelByID(ctx, ch.ID); err != ErrNotFound {
		t.Fatalf("删服后频道应消失，实际 %v", err)
	}

	// invites 用尽。
	inv := &Invite{Code: "abc", ServerID: 99, CreatorID: 1, MaxUses: 1}
	// 手动注入一个 server 供 invite 外键语义。
	s2 := &Server{Name: "s2", OwnerID: 1}
	_ = m.CreateServer(ctx, s2)
	inv.ServerID = s2.ID
	if err := m.CreateInvite(ctx, inv); err != nil {
		t.Fatal(err)
	}
	if err := m.UseInvite(ctx, "abc"); err != nil {
		t.Fatal(err)
	}
	if err := m.UseInvite(ctx, "abc"); err != ErrConflict {
		t.Fatalf("用尽应 ErrConflict，实际 %v", err)
	}

	// KV：Incr 语义。
	n, err := m.Incr(ctx, "k", 0)
	if err != nil || n != 1 {
		t.Fatalf("Incr 首次 = %d %v", n, err)
	}
	n, _ = m.Incr(ctx, "k", 0)
	if n != 2 {
		t.Fatalf("Incr 第二次 = %d", n)
	}
}
