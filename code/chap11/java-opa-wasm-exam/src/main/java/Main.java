import io.github.sangkeon.opa.wasm.OPAModule;

public class Main {
    public static void main(String[] args) {
        try (
            OPAModule om = new OPAModule("./sample-policy/wasm/policy.wasm");
        ) {
            String input = "{\"user\": \"john\"}";
            String data = "{\"role\":{\"john\":\"admin\"}}";

            om.setData(data);

            String output = om.evaluate(input);
            System.out.println("Result=" + output);
        }
    }
}