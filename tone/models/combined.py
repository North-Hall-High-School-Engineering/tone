import torch.nn as nn

from models.regressor import RegressorHead


class CombinedModel(nn.Module):
    def __init__(self, hubert_model):
        super().__init__()
        self.model = CombinedInferenceModel(hubert_model)
        self.loss_fn = nn.MSELoss()

    def forward(self, input_values, attention_mask=None, labels=None):
        preds = self.model(input_values, attention_mask)

        loss = None
        if labels is not None:
            loss = self.loss_fn(preds, labels)

        return {
            "loss": loss,
            "logits": preds,
        }


class CombinedInferenceModel(nn.Module):
    def __init__(self, hubert_model):
        super().__init__()
        self.hubert = hubert_model
        self.regressor = RegressorHead(hubert_model.config.hidden_size)

    def forward(self, input_values, attention_mask):
        outputs = self.hubert(
            input_values=input_values,
            attention_mask=attention_mask,
        )

        preds = self.regressor(outputs.last_hidden_state)
        return preds
