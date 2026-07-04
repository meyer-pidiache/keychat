export type NostrEvent = {
  id: string
  pubkey: string
  created_at: number
  kind: number
  tags: string[][]
  content: string
  sig: string
}

export type SubscriptionCallback = (event: NostrEvent) => void
export type NoticeCallback = (message: string) => void
export type ConnectionCallback = (connected: boolean) => void

export class RelayClient {
  private ws: WebSocket | null = null
  private url: string
  private subscriptions = new Map<string, SubscriptionCallback>()
  private noticeCb: NoticeCallback | null = null
  private connectionCb: ConnectionCallback | null = null
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null
  private shouldReconnect = false

  constructor(url: string) {
    this.url = url
  }

  connect(): void {
    if (this.ws && (this.ws.readyState === WebSocket.OPEN || this.ws.readyState === WebSocket.CONNECTING)) {
      return
    }

    this.shouldReconnect = true

    try {
      this.ws = new WebSocket(this.url)
    } catch {
      this.handleConnectionChange(false)
      return
    }

    this.ws.onopen = () => {
      this.handleConnectionChange(true)
    }

    this.ws.onclose = () => {
      this.handleConnectionChange(false)
      if (this.shouldReconnect) {
        this.scheduleReconnect()
      }
    }

    this.ws.onerror = () => {
      this.handleConnectionChange(false)
    }

    this.ws.onmessage = (event: MessageEvent) => {
      try {
        const data = JSON.parse(event.data)
        this.handleMessage(data)
      } catch {
      }
    }
  }

  disconnect(): void {
    this.shouldReconnect = false
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
    if (this.ws) {
      this.ws.onclose = null
      this.ws.close()
      this.ws = null
    }
    this.handleConnectionChange(false)
  }

  isConnected(): boolean {
    return this.ws !== null && this.ws.readyState === WebSocket.OPEN
  }

  subscribe(subId: string, filter: Record<string, unknown>, cb: SubscriptionCallback): void {
    this.subscriptions.set(subId, cb)
    this.send(['REQ', subId, filter])
  }

  unsubscribe(subId: string): void {
    this.subscriptions.delete(subId)
    this.send(['CLOSE', subId])
  }

  publish(event: NostrEvent): void {
    this.send(['EVENT', event])
  }

  onNotice(cb: NoticeCallback): void {
    this.noticeCb = cb
  }

  onConnectionChange(cb: ConnectionCallback): void {
    this.connectionCb = cb
  }

  private send(msg: unknown[]): void {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(msg))
    }
  }

  private handleMessage(data: unknown): void {
    if (!Array.isArray(data) || data.length < 2) return

    const [type, ...rest] = data as [string, ...unknown[]]

    switch (type) {
      case 'EVENT': {
        const subId = rest[0] as string
        const event = rest[1] as NostrEvent
        const cb = this.subscriptions.get(subId)
        if (cb) cb(event)
        break
      }
      case 'EOSE': {
        break
      }
      case 'NOTICE': {
        const msg = rest[0] as string
        if (this.noticeCb) this.noticeCb(msg)
        break
      }
      case 'OK': {
        break
      }
    }
  }

  private scheduleReconnect(): void {
    if (this.reconnectTimer) return
    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null
      if (this.shouldReconnect) {
        this.connect()
      }
    }, 3000)
  }

  private handleConnectionChange(connected: boolean): void {
    if (this.connectionCb) {
      this.connectionCb(connected)
    }
  }
}
