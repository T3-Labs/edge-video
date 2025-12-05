import pika
import sys
import cv2
import numpy as np
from datetime import datetime
from collections import deque
import time

# Configurações do RabbitMQ
RABBITMQ_HOST = '34.71.212.239'
RABBITMQ_PORT = 5672
RABBITMQ_VHOST = 'supercarlao_rj_mercado'
RABBITMQ_USER = 'supercarlao_rj_mercado'
RABBITMQ_PASS = 'vTUaGG9S0Kbl_1AhqPxudML1jCYHRq0efkRDA6p_lgM'

# Pega ID da câmera do argumento da linha de comando (padrão: cam1)
CAMERA_ID = sys.argv[1] if len(sys.argv) > 1 else 'cam1'

EXCHANGE_NAME = 'supercarlao_rj_mercado.exchange'
ROUTING_KEY = f'supercarlao_rj_mercado.{CAMERA_ID}'

# Mantém apenas os 3 frames mais recentes
frame_buffer = deque(maxlen=3)
frame_count = 0
leak_count = 0  # Contador de vazamentos (frames de outras câmeras)
strip_count = 0  # Contador de frames com metadata removido
decode_errors = 0  # Contador de erros de decodificação
pil_rescues = 0  # Contador de frames salvos pelo decoder PIL
last_display_time = 0
fps_target = 15  # 15 FPS = 66.67ms por frame
frame_interval = 1.0 / fps_target
DEBUG_STRIP = False  # Ativa debug do strip de metadata (desabilitado em produção)

def on_frame_received(ch, method, properties, body):
    """Callback para receber frames"""
    global frame_count, frame_buffer, leak_count, strip_count, decode_errors, pil_rescues

    frame_count += 1

    # 🔍 VALIDAÇÃO TRIPLA: Routing Key + Header + Resolução
    received_routing_key = method.routing_key

    # DEBUG ULTRA-DETALHADO: Primeiros 10 frames
    if frame_count <= 10:
        header_cam_id = "N/A"
        if properties.headers and 'camera_id' in properties.headers:
            header_cam_id = properties.headers['camera_id']
            if isinstance(header_cam_id, bytes):
                header_cam_id = header_cam_id.decode('utf-8')

        # Mostra primeiros 16 bytes do frame
        frame_fingerprint = body[:16].hex() if len(body) >= 16 else body.hex()
        print(f"[RECV #{frame_count}] RoutingKey={received_routing_key}, Header[camera_id]={header_cam_id}, Size={len(body)}, Fingerprint={frame_fingerprint}")

    # Validação 1: Routing Key
    if received_routing_key != ROUTING_KEY:
        leak_count += 1
        print(f"[VAZAMENTO ROUTING] Frame #{frame_count} - Esperado: {ROUTING_KEY}, Recebido: {received_routing_key} (Tamanho: {len(body)} bytes)")
        return  # Ignora frame de outra câmera

    # Validação 2: Header camera_id (se disponível)
    if properties.headers and 'camera_id' in properties.headers:
        header_cam_id = properties.headers['camera_id']
        if isinstance(header_cam_id, bytes):
            header_cam_id = header_cam_id.decode('utf-8')

        if header_cam_id != CAMERA_ID:
            leak_count += 1
            print(f"[VAZAMENTO HEADER] Frame #{frame_count} - Esperado: {CAMERA_ID}, Recebido: {header_cam_id} (Tamanho: {len(body)} bytes)")
            return  # Ignora frame de outra câmera

    try:
        # Remove metadata FFmpeg (FFFE comment marker) se presente
        cleaned_body = body
        stripped = False

        if len(body) >= 6 and body[0:2] == b'\xff\xd8' and body[2:4] == b'\xff\xfe':
            # JPEG com comment marker (FFFE)
            # Formato: FFD8 FFFE LLLL [comment data] [resto do JPEG]
            # Length field: big-endian, INCLUI os 2 bytes do próprio length field
            comment_len = (body[4] << 8) | body[5]  # Big-endian length

            # Total a pular desde DEPOIS do SOI (FFD8):
            # - Marker FFFE: 2 bytes
            # - Length + Data: comment_len bytes (já inclui os 2 bytes do length)
            skip_from_soi = 2 + comment_len  # Skip desde posição 2 (após FFD8)

            # Posição final: 2 (SOI) + skip_from_soi
            next_marker_pos = 2 + skip_from_soi

            if next_marker_pos < len(body):
                # Reconstrói JPEG: SOI + resto após comment
                cleaned_body = b'\xff\xd8' + body[next_marker_pos:]
                stripped = True
                strip_count += 1

                if DEBUG_STRIP and frame_count <= 3:
                    print(f"[STRIP DEBUG] Frame #{frame_count}")
                    print(f"  Comment len field: {comment_len} (0x{comment_len:04x})")
                    print(f"  Next marker at: {next_marker_pos}")
                    print(f"  Next bytes: {body[next_marker_pos:next_marker_pos+4].hex()}")
                    print(f"  Before size: {len(body)}, After size: {len(cleaned_body)}")

                # 🔬 VERIFICAÇÃO DE INTEGRIDADE: Descarta frames sem EOF (corrompidos)
                has_eof = cleaned_body[-2:] == b'\xff\xd9'

                # Debug para primeiros 5 frames
                if DEBUG_STRIP and frame_count <= 5:
                    eof_status = "✓ EOF OK" if has_eof else "✗ SEM EOF"
                    eof_search = body[-20:].find(b'\xff\xd9')
                    if eof_search >= 0:
                        eof_pos = len(body) - 20 + eof_search
                        padding_bytes = len(body) - eof_pos - 2
                        eof_info = f"EOF em -{20-eof_search} bytes ({padding_bytes} bytes padding)"
                    else:
                        eof_info = "EOF NÃO ENCONTRADO nos últimos 20 bytes"
                    print(f"[INTEGRITY #{frame_count}] Size={len(body)}, {eof_status}, {eof_info}, Last4={body[-4:].hex()}")

                # Ignora frames corrompidos (sem EOF) ANTES de tentar decodificar
                if not has_eof:
                    decode_errors += 1
                    if frame_count <= 10:  # Loga só os primeiros 10
                        print(f"[SKIP] Frame #{frame_count} corrompido (sem EOF) - descartando")
                    return  # Retorna sem processar
            else:
                # Se skip_bytes inválido, usa original
                if DEBUG_STRIP:
                    print(f"[STRIP ERROR] Frame #{frame_count} - Invalid next_marker_pos: {next_marker_pos} >= {len(body)}")

        # Decodifica o JPEG - tenta OpenCV primeiro, fallback para PIL
        np_arr = np.frombuffer(cleaned_body, np.uint8)
        img = cv2.imdecode(np_arr, cv2.IMREAD_COLOR)

        # Se OpenCV falhou, tenta PIL/Pillow (mais tolerante a JPEGs corrompidos)
        if img is None:
            try:
                from PIL import Image
                import io
                img_pil = Image.open(io.BytesIO(cleaned_body))
                img = np.array(img_pil)
                img = cv2.cvtColor(img, cv2.COLOR_RGB2BGR)
                pil_rescues += 1  # Contabiliza resgate pelo PIL
            except Exception as e:
                # PIL também falhou, frame realmente corrompido
                pass

        if img is not None:
            # VALIDAÇÃO CRÍTICA: Verifica se resolução está consistente
            height, width = img.shape[:2]

            # Define resolução esperada (ajuste conforme suas câmeras)
            # Se primeira vez, salva como referência
            if frame_count == 1:
                globals()['EXPECTED_WIDTH'] = width
                globals()['EXPECTED_HEIGHT'] = height
                print(f"[RESOLUÇÃO REFERÊNCIA] {width}x{height}")

            # Se resolução mudou, DESCARTA frame suspeito
            if 'EXPECTED_WIDTH' in globals():
                expected_w = globals()['EXPECTED_WIDTH']
                expected_h = globals()['EXPECTED_HEIGHT']

                # Tolerância de 10% para pequenas variações
                width_diff = abs(width - expected_w) / expected_w
                height_diff = abs(height - expected_h) / expected_h

                if width_diff > 0.1 or height_diff > 0.1:
                    decode_errors += 1
                    print(f"[RESOLUÇÃO INVÁLIDA] Frame #{frame_count} - Esperado: {expected_w}x{expected_h}, Recebido: {width}x{height} - DESCARTANDO")
                    return  # Descarta frame com resolução errada

            # Pega o timestamp do AMQP (se disponível)
            timestamp = properties.timestamp if properties.timestamp else time.time()

            # Adiciona ao buffer com timestamp
            frame_buffer.append({
                'image': img,
                'timestamp': timestamp,
                'count': frame_count,
                'size': len(body),
                'routing_key': received_routing_key
            })

        else:
            decode_errors += 1
            print(f"[ERRO] Frame #{frame_count} não decodificou - Routing: {received_routing_key}, Tamanho: {len(body)} bytes, Header: {body[:10].hex() if len(body) >= 10 else 'vazio'}, Stripped: {stripped}")

    except Exception as e:
        decode_errors += 1
        print(f"[ERRO] Frame #{frame_count} - Exceção: {e}, Routing: {received_routing_key}, Tamanho: {len(body)} bytes")

def main():
    global last_display_time

    # Mostra help se solicitado
    if len(sys.argv) > 1 and sys.argv[1] in ['-h', '--help', 'help']:
        print("="*80)
        print("EDGE VIDEO VIEWER - Visualizador de Câmeras")
        print("="*80)
        print("\nUso:")
        print("  python viewer_cam1_sync.py [CAMERA_ID]")
        print("\nExemplos:")
        print("  python viewer_cam1_sync.py           # Visualiza cam1 (padrão)")
        print("  python viewer_cam1_sync.py cam1      # Visualiza cam1")
        print("  python viewer_cam1_sync.py cam2      # Visualiza cam2")
        print("  python viewer_cam1_sync.py cam3      # Visualiza cam3")
        print("  python viewer_cam1_sync.py cam4      # Visualiza cam4")
        print("  python viewer_cam1_sync.py cam5      # Visualiza cam5")
        print("  python viewer_cam1_sync.py cam6      # Visualiza cam6")
        print("\nCâmeras disponíveis no config.yaml:")
        print("  - cam1 (Mercado Autônomo - RTMP)")
        print("  - cam2 (Pix Force Canal 1 - RTSP)")
        print("  - cam3 (Pix Force Canal 2 - RTSP)")
        print("  - cam4 (Pix Force Canal 3 - RTSP)")
        print("  - cam5 (Pix Force Canal 4 - RTSP)")
        print("  - cam6 (Pix Force Canal 5 - RTSP)")
        print("="*80)
        sys.exit(0)

    print("="*80)
    print(f"VISUALIZADOR SINCRONIZADO - {CAMERA_ID}")
    print("="*80)
    print(f"RabbitMQ: {RABBITMQ_HOST}:{RABBITMQ_PORT}")
    print(f"VHost: {RABBITMQ_VHOST}")
    print(f"Exchange: {EXCHANGE_NAME}")
    print(f"Routing Key: {ROUTING_KEY}")
    print(f"Target FPS: {fps_target}")
    print("="*80)
    print("\nPressione 'q' na janela de vídeo ou CTRL+C para sair\n")

    credentials = pika.PlainCredentials(RABBITMQ_USER, RABBITMQ_PASS)
    parameters = pika.ConnectionParameters(
        host=RABBITMQ_HOST,
        port=RABBITMQ_PORT,
        virtual_host=RABBITMQ_VHOST,
        credentials=credentials,
        heartbeat=600,
        blocked_connection_timeout=300
    )

    try:
        connection = pika.BlockingConnection(parameters)
        channel = connection.channel()

        print("[OK] Conectado ao RabbitMQ!\n")

        # Configura consumo com prefetch=1 para não acumular
        channel.basic_qos(prefetch_count=1)
        channel.exchange_declare(exchange=EXCHANGE_NAME, exchange_type='topic', durable=True)
        result = channel.queue_declare(queue='', exclusive=True)
        queue_name = result.method.queue
        channel.queue_bind(exchange=EXCHANGE_NAME, queue=queue_name, routing_key=ROUTING_KEY)
        channel.basic_consume(queue=queue_name, on_message_callback=on_frame_received, auto_ack=True)

        print(f"[OK] Aguardando frames da {CAMERA_ID}...\n")

        last_display_time = time.time()
        frames_displayed = 0
        start_time = time.time()

        # Loop de consumo e display
        while True:
            try:
                # Processa mensagens do RabbitMQ (não bloqueante)
                channel.connection.process_data_events(time_limit=0.01)

                # Controle de FPS - exibe apenas se passou tempo suficiente
                current_time = time.time()
                time_since_last = current_time - last_display_time

                if time_since_last >= frame_interval and len(frame_buffer) > 0:
                    # Pega o frame mais recente do buffer
                    latest_frame = frame_buffer[-1]

                    img = latest_frame['image']

                    # Redimensiona para 50% (640x360)
                    height, width = img.shape[:2]
                    resized = cv2.resize(img, (width // 2, height // 2))

                    # Calcula FPS real
                    elapsed = current_time - start_time
                    fps_real = frames_displayed / elapsed if elapsed > 0 else 0

                    # OVERLAY GRANDE: ID da câmera (centro-topo) para evitar confusão
                    cam_text = f'CAMERA: {CAMERA_ID.upper()}'
                    text_size = cv2.getTextSize(cam_text, cv2.FONT_HERSHEY_DUPLEX, 1.2, 3)[0]
                    text_x = (resized.shape[1] - text_size[0]) // 2
                    # Fundo semi-transparente para o texto
                    overlay = resized.copy()
                    cv2.rectangle(overlay, (text_x - 10, 5), (text_x + text_size[0] + 10, 45), (0, 0, 0), -1)
                    cv2.addWeighted(overlay, 0.6, resized, 0.4, 0, resized)
                    cv2.putText(resized, cam_text, (text_x, 35),
                               cv2.FONT_HERSHEY_DUPLEX, 1.2, (0, 255, 255), 3)

                    # Adiciona informações laterais
                    cv2.putText(resized, f'Frame #{latest_frame["count"]}', (5, 60),
                               cv2.FONT_HERSHEY_SIMPLEX, 0.4, (0, 255, 0), 1)
                    cv2.putText(resized, f'FPS: {fps_real:.1f}', (5, 80),
                               cv2.FONT_HERSHEY_SIMPLEX, 0.4, (0, 255, 0), 1)
                    cv2.putText(resized, f'Buffer: {len(frame_buffer)}', (5, 100),
                               cv2.FONT_HERSHEY_SIMPLEX, 0.4, (0, 255, 0), 1)
                    cv2.putText(resized, datetime.now().strftime('%H:%M:%S'), (5, 120),
                               cv2.FONT_HERSHEY_SIMPLEX, 0.4, (0, 255, 0), 1)

                    # Exibe o frame
                    cv2.imshow(f'Edge Video - {CAMERA_ID} (Sync)', resized)

                    last_display_time = current_time
                    frames_displayed += 1

                    # Log periódico
                    if frames_displayed % 30 == 0:
                        print(f"[INFO] Exibidos {frames_displayed} frames, FPS real: {fps_real:.1f}, "
                              f"Buffer size: {len(frame_buffer)}, Recebidos: {frame_count}")

                # Processa eventos do OpenCV
                if cv2.waitKey(1) & 0xFF == ord('q'):
                    print("\n[INFO] Saindo...")
                    break

            except KeyboardInterrupt:
                print('\n[INFO] Interrompido pelo usuário')
                break

        cv2.destroyAllWindows()
        connection.close()

        # Estatísticas finais
        total_time = time.time() - start_time
        correct_frames = frame_count - leak_count
        print(f"\n{'='*80}")
        print(f"ESTATÍSTICAS:")
        print(f"  - Frames TOTAIS recebidos: {frame_count}")
        print(f"  - Frames CORRETOS (cam esperada): {correct_frames}")
        print(f"  - Frames VAZADOS (outras câmeras): {leak_count} ({(leak_count/frame_count*100):.1f}%)" if frame_count > 0 else "")
        print(f"  - Frames com metadata removido: {strip_count} ({(strip_count/frame_count*100):.1f}%)" if frame_count > 0 else "")
        print(f"  - Frames salvos por PIL/Pillow: {pil_rescues} ({(pil_rescues/frame_count*100):.1f}%)" if frame_count > 0 else "")
        print(f"  - Erros de decodificação: {decode_errors} ({(decode_errors/frame_count*100):.1f}%)" if frame_count > 0 else "")
        print(f"  - Frames exibidos: {frames_displayed}")
        print(f"  - Tempo total: {total_time:.1f}s")
        print(f"  - FPS médio: {frames_displayed/total_time:.1f}" if total_time > 0 else "  - FPS médio: N/A")
        print(f"  - Taxa de descarte: {((correct_frames - frames_displayed)/correct_frames*100):.1f}%" if correct_frames > 0 else "  - Taxa de descarte: N/A")
        print(f"{'='*80}")

        # Alerta se houve vazamento
        if leak_count > 0:
            print(f"\n⚠️  ATENÇÃO: Detectados {leak_count} frames de OUTRAS câmeras!")
            print(f"   Isso indica VAZAMENTO no Edge Video Publisher!")
            print(f"   O publisher está enviando frames com routing keys erradas.")
            print(f"{'='*80}\n")

    except pika.exceptions.AMQPConnectionError as e:
        print(f"\n[ERRO] Erro de conexão: {e}")
    except Exception as e:
        print(f"\n[ERRO] Erro inesperado: {e}")
        import traceback
        traceback.print_exc()

if __name__ == '__main__':
    main()
