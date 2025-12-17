package queue

import "cobranca-bmp/helpers"

// Declarando as filas necessárias para o processamento de CreditoPessoal
func (r *RabbitMQ) declareCobrancaQueues() error {

	_, err := r.cobrancaCh.QueueDeclare(r.cobrancaQueue, true, false, false, false, nil)
	if err != nil {
		return err
	}
	err = r.cobrancaCh.QueueBind(r.cobrancaQueue, r.cobrancaQueue, r.exchange, false, nil)
	if err != nil {
		return err
	}

	return nil

}

// Obtendo os canais da conexão, eles serão utilizados pelos consumers.
func (r *RabbitMQ) setUpCobrancaChannels() error {
	ch, err := r.conn.Channel()
	if err != nil {
		return err
	}

	r.cobrancaCh = ch

	return nil
}

// Fechando os canais da conexão, eles serão utilizados pelos consumers.
func (r *RabbitMQ) closeCobranca() error {

	if err := r.cobrancaCh.Close(); err != nil {
		helpers.LogError(r.ctx, r.logger, r.location, "rabbitmq-close", "", "falha ao fechar canal de cobrança", err.Error(), nil)
		return err
	}
	helpers.LogInfo(r.ctx, r.logger, r.location, "rabbitmq-close", "", "canal de cobrança fechado", nil)

	return nil

}

// checkCreditoPessoal verifica se os canais de CreditoPessoal estão abertos.
func (r *RabbitMQ) checkCreditoPessoal() (map[string]any, error) {
	var info = make(map[string]any)
	var errors = make([]string, 0)

	/*if r.conn.IsClosed() {
		errors = append(errors, "conexão fechada com o Rabbitmq")
	} */

	if r.cobrancaCh.IsClosed() {
		errors = append(errors, "canal de cobrança fechado")
		info["Cobranca"] = "fechado"
	} else {
		info["Cobranca"] = "ok"
	}

	if len(errors) > 0 {
		info["errors"] = errors
	} else {
		info["status"] = "ok"
		//info["info"] = r.conn.ConnectionState()
	}
	return info, nil

}
