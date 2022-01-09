import io.github.sangkeon.opa.wasm.Bundle;
import io.github.sangkeon.opa.wasm.BundleUtil;
import io.github.sangkeon.opa.wasm.OPAModule;

public class BundleMain {
    public static void main(String[] args) {
        try {
            Bundle bundle = BundleUtil.extractBundle("./sample-policy/bundle/bundle.tar.gz");

            try (
                OPAModule om = new OPAModule(bundle);
            ) {
                String input = "{\"user\": \"alice\"}";
                String output = om.evaluate(input);

                System.out.println("Result=" + output);
           }

        } catch(Exception e) {
            e.printStackTrace();
        }
    }
}